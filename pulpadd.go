/* Pulp CLI
 *
 * - Version 1.1.1 - 2021/08/27
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

func pulpCreateArtifact(pack string) (PulpCreate, error) {

	var pc = PulpCreate{
		Pulp_href:    "",
		Pulp_created: "",
		File:         "",
		Size:         0,
		Md5:          "",
		Sha1:         "",
		Sha224:       "",
		Sha256:       "",
		Sha384:       "",
		Sha512:       "",
	}

	fileDir, err := os.Getwd()
	if err != nil {
		return pc, err
	}
	filePath := path.Join(fileDir, pack)
	file, err := os.Open(filePath)
	if err != nil {
		return pc, err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return pc, err
	}
	io.Copy(part, file)
	writer.Close()
	req, err := http.NewRequest("POST", apiEnd+"/artifacts/", body)
	if err != nil {
		return pc, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	result, status, err := pulpExec(req)
	if err != nil {
		return pc, err
	}
	if status != http.StatusCreated {
		return pc, fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
	}
	err = json.Unmarshal(result, &pc)
	if err != nil {
		return pc, err
	}
	return pc, nil
}

func pulpAddArtifactToContents(details PulpCreate, pack string) ([]string, error) {

	artifact := Artifact{
		Artifact: details.Pulp_href,
		Rel_path: filepath.Base(pack),
	}
	body, err := json.Marshal(artifact)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("POST", apiEnd+"/content/rpm/packages/", data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	result, status, err := pulpExec(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusAccepted {
		return nil, fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
	}
	decoder := json.NewDecoder(bytes.NewReader(result))
	task := Task{}
	err = decoder.Decode(&task)
	if err != nil {
		return nil, err
	}
	taskResults, err := pulpWaitForTask(task)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Artifact added to Contents.\n")
	return taskResults.Created_resources, nil
}

func pulpAddContentsToRepo(repo string, resources []string) error {

	r, err := pulpRepositoryInfo(repo)
	if err != nil {
		return err
	}
	if r.Count == 0 {
		return fmt.Errorf("repository %s does not exist", repo)
	}
	repoResults := r.Results[0]
	content := AddContentUnits{
		Add_content_units: resources,
	}
	body, err := json.Marshal(content)
	if err != nil {
		return err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("POST", apiSrv+repoResults.Pulp_href+"modify/", data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	result, status, err := pulpExec(req)
	if err != nil {
		return err
	}
	if status != http.StatusAccepted {
		return fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
	}
	decoder := json.NewDecoder(bytes.NewReader(result))
	task := Task{}
	err = decoder.Decode(&task)
	if err != nil {
		return err
	}
	_, err = pulpWaitForTask(task)
	if err != nil {
		return err
	}

	fmt.Printf("Content added to repository.\n")
	return nil
}

func pulpInitUpload(size int64) (PulpUploadResults, error) {

	var (
		upload = UploadStart{
			Size: 0,
		}
		pur = PulpUploadResults{
			Pulp_href:    "",
			Pulp_created: "",
			Size:         0,
			Completed:    "",
		}
	)

	upload.Size = size
	body, err := json.Marshal(upload)
	if err != nil {
		return pur, err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("POST", apiEnd+"/uploads/", data)
	if err != nil {
		return pur, err
	}
	req.Header.Set("Content-Type", "application/json")
	result, status, err := pulpExec(req)
	if err != nil {
		return pur, err
	}
	if status != http.StatusCreated {
		return pur, fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
	}
	err = json.Unmarshal(result, &pur)
	if err != nil {
		return pur, err
	}
	return pur, nil
}

func pulpDeinitUpload(pur PulpUploadResults) error {

	req, err := http.NewRequest("DELETE", apiSrv+pur.Pulp_href, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, status, err := pulpExec(req)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("HTTP response: %d, expected: %d", status, http.StatusNoContent)
	}
	return nil
}

func pulpFinishUpload(pur PulpUploadResults, pack string) ([]string, error) {

	var (
		finish = UploadFinish{
			Sha256: "",
		}
		task = Task{
			Task: "",
		}
	)

	sha256, err := calcSHA256(pack)
	if err != nil {
		return nil, err
	}
	finish.Sha256 = sha256
	body, err := json.Marshal(finish)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("POST", apiSrv+pur.Pulp_href+"commit/", data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	result, status, err := pulpExec(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusAccepted {
		return nil, fmt.Errorf("HTTP response: %d, expected: %d", status, http.StatusAccepted)
	}
	err = json.Unmarshal(result, &task)
	if err != nil {
		return nil, err
	}
	taskResults, err := pulpWaitForTask(task)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Artifact created and uploaded chunks removed.\n")
	return taskResults.Created_resources, nil
}

func pulpChunkThread(c chan int64, t *Thread, f string, s int64, p PulpUploadResults) {

	// c = channel, t = thread, f = file, s = size, p = pulpUploadResults
	var (
		offset       int64
		ok           bool
		bytesread    int
		byteswritten int64
		temp         string
	)

	// Each thread (goroutine) needs its own filedescriptor in order to
	// read in parallel.
	file, err := os.Open(f)
	if err != nil {
		t.mutex.Lock()
		t.count--
		// Checking for nil ensures only the first error is logged.
		if t.err == nil {
			t.err = err
		}
		t.mutex.Unlock()
		return
	}
	defer file.Close()
	// Each thread (goroutine) needs a dedicated http connection, so that
	// parallel communication is possible.
	threadClient := &http.Client{Timeout: time.Second * CLIENT_TIMEOUT}
	defer threadClient.CloseIdleConnections()
	// A read/write buffer for processing chunks.
	buffer := make([]byte, CHUNKSIZE)
	for {
		offset, ok = <-c
		if ok {
			seek, err := file.Seek(offset, 0)
			if err != nil {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			if seek != offset {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = fmt.Errorf("seek offset %d in %s expected offset %d", seek, f, offset)
				}
				t.mutex.Unlock()
				break
			}
			// Reset buffer to its maximum size. In multi-threaded operation,
			// the buffer might have been shrunk to accommodate the read of
			// the last part of the file. But it is possible for a non-last
			// chunk to exist that still needs to be send, so buffer must be
			// reset!
			buffer = buffer[0:CHUNKSIZE]
			bytesread, err = file.Read(buffer)
			if err != nil {
				if err != io.EOF {
					t.mutex.Lock()
					t.count--
					if t.err == nil {
						t.err = err
					}
					t.mutex.Unlock()
					break
				} else {
					err = nil
				}
			}
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			temp = fmt.Sprintf("chunk_offset_%d", offset)
			part, err := writer.CreateFormFile("file", temp)
			if err != nil {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			// io.Copy will copy the complete contents of buffer. However the buffer is not
			// always filled to its maximum size. The last chunk of a file is usually smaller
			// than the chunksize. As a result garbage still in the buffer will also be copied.
			// To prohibit this, we explicitly need to set the buffer lenght.
			buffer = buffer[0:bytesread]
			byteswritten, err = io.Copy(part, bytes.NewReader(buffer))
			if err != nil {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			if int64(bytesread) != byteswritten {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = fmt.Errorf("%d bytes read, but %d bytes written", bytesread, byteswritten)
				}
				t.mutex.Unlock()
				break
			}
			err = writer.Close()
			if err != nil {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			request, err := http.NewRequest("PUT", apiSrv+p.Pulp_href, body)
			if err != nil {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			temp = fmt.Sprintf("bytes %d-%d/%d", offset, offset+int64(bytesread)-1, s)
			request.Header.Add("Content-Type", writer.FormDataContentType())
			request.Header.Add("Content-Range", temp)
			request.SetBasicAuth(apiUser, apiPass)
			response, err := threadClient.Do(request)
			if err != nil {
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			_, err = ioutil.ReadAll(response.Body)
			if err != nil {
				response.Body.Close()
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = err
				}
				t.mutex.Unlock()
				break
			}
			if response.StatusCode != http.StatusOK {
				response.Body.Close()
				t.mutex.Lock()
				t.count--
				if t.err == nil {
					t.err = fmt.Errorf("HTTP response: %d, expected: %d", response.StatusCode, http.StatusOK)
				}
				t.mutex.Unlock()
				break
			}
			response.Body.Close()
			fmt.Printf("chunk %s uploaded\n", temp)
		} else {
			t.mutex.Lock()
			t.count--
			t.mutex.Unlock()
			break
		}
	}
}

func pulpUploadChunks(file string, size int64, pur PulpUploadResults) error {

	var (
		offset int64
		t      Thread
	)

	c := make(chan int64, MAX_THREADS)
	for i := 0; i < MAX_THREADS; i++ {
		go pulpChunkThread(c, &t, file, size, pur)
		t.mutex.Lock()
		t.count++
		t.mutex.Unlock()
	}
	for offset = 0; offset < size; offset += CHUNKSIZE {
		t.mutex.Lock()
		if t.err != nil {
			t.mutex.Unlock()
			break
		}
		t.mutex.Unlock()
		c <- offset
	}
	close(c) // We're done sending chunks, inform threads to terminate.
	t.mutex.Lock()
	for t.count != 0 {
		t.mutex.Unlock()
		time.Sleep(time.Second)
		t.mutex.Lock()
	}
	t.mutex.Unlock()
	return t.err
}

func pulpAddPackage(repo, pack string) error {

	var (
		pc = PulpCreate{
			Pulp_href:    "",
			Pulp_created: "",
			File:         "",
			Size:         0,
			Md5:          "",
			Sha1:         "",
			Sha224:       "",
			Sha256:       "",
			Sha384:       "",
			Sha512:       "",
		}
		resources []string
	)

	results, err := pulpPackageInfo(pack)
	if err != nil {
		return err
	}
	if results.Count > 0 {
		return fmt.Errorf("package already exists")
	}
	fileDir, err := os.Getwd()
	if err != nil {
		return err
	}
	file := path.Join(fileDir, pack)
	size, err := getFileSize(file)
	if err != nil {
		return err
	}
	/*
	 * If the package size is less than CHUNKSIZE, we'll
	 * perform a direct upload. If not, a chunked upload
	 * is performed.
	 */
	if size > CHUNKSIZE {
		pur, err := pulpInitUpload(size)
		if err != nil {
			return err
		}
		err = pulpUploadChunks(file, size, pur)
		if err != nil {
			err2 := pulpDeinitUpload(pur)
			if err2 != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				return err2
			} else {
				return err
			}
		}
		resources, err = pulpFinishUpload(pur, pack)
		if err != nil {
			err2 := pulpDeinitUpload(pur)
			if err2 != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				return err2
			} else {
				return err
			}
		}
		pc, err = pulpArtifactInfo(resources[0])
		if err != nil {
			return err
		}
	} else {
		pc, err = pulpCreateArtifact(pack)
		if err != nil {
			return err
		}
	}
	resources, err = pulpAddArtifactToContents(pc, pack)
	if err != nil {
		return err
	}
	err = pulpAddContentsToRepo(repo, resources)
	if err != nil {
		return err
	}
	return nil
}
