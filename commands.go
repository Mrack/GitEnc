/*
 * Copyright (c) 2023 Mrack
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 * This program is named gitenc and is distributed under the terms of
 * the GNU General Public License, version 3 or any later version.
 */

package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	log "gitenc/log"
	"io/ioutil"
	"os"
	"strings"
	"unsafe"
)

type HEADER struct {
	flag    [4]byte
	version byte
	khash   [16]byte
	fhash   [16]byte
	size    uint64
}

func blobIsEncrypted(blob string) bool {
	_, output := RunCommand("git", "cat-file", "blob", blob)
	data := []byte(output)
	if len(data) > int(unsafe.Sizeof(HEADER{})) {
		header := (*HEADER)(unsafe.Pointer(&data[0]))
		if header.flag == [4]byte{0, 'M', 'R', 0} && header.version == 1 {
			return true
		}
	}
	return false
}
func ClearGitConfig(name string) {
	ex, _ := os.Executable()
	ex = "'" + ex + "'"
	// git config --unset filter.gitenc.clean
	RunCommand("git", "config", "--unset", "filter.gitenc.clean")
	// git config --unset filter.gitenc.smudge
	RunCommand("git", "config", "--unset", "filter.gitenc.smudge")
	// git config --unset filter.gitenc.required
	RunCommand("git", "config", "--unset", "filter.gitenc.required")
	// git config --unset diff.gitenc.textconv
	RunCommand("git", "config", "--unset", "diff.gitenc.textconv")
}
func SetGitConfig(name string) {
	ex, _ := os.Executable()
	ex = "'" + ex + "'"
	// git config filter.gitenc.smudge "gitenc smudge %f"
	RunCommand("git", "config", "filter.gitenc.smudge", ex+" smudge -keyname "+name)
	// git config filter.gitenc.clean "gitenc clean %f"
	RunCommand("git", "config", "filter.gitenc.clean", ex+" clean -keyname "+name)
	// git config filter.gitenc.required true
	RunCommand("git", "config", "filter.gitenc.required", "true")
	// git config diff.gitenc.textconv "gitenc diff %f"
	RunCommand("git", "config", "diff.gitenc.textconv", ex+" diff -keyname "+name)
}
func getEncryptFiles() []string {
	//git ls-files -cz -- .
	_, output := RunCommand("git", "ls-files", "-cz", "--", getRepoRoot())
	fileList := bytes.Split([]byte(output), []byte("\000"))
	encrypted := make([]string, 0)
	for _, fileInfo := range fileList {
		if len(fileInfo) == 0 {
			continue
		}
		fields := strings.Fields(string(fileInfo))
		if fields[0] == "?" {
			continue
		}
		// git check-attr filter diff -- filename
		_, output := RunCommand("git", "check-attr", "filter", "diff", "--", fields[0])
		attrs := strings.Fields(string(output))
		filter, diff := attrs[2], attrs[5]

		if filter == "gitenc" && diff == "gitenc" {
			encrypted = append(encrypted, fields[0])
		}
	}
	return encrypted
}

func getRepoRoot() string {
	_, res := RunCommand("git", "rev-parse", "--show-cdup")
	res = Trim(res)
	if res == "" {
		return "."
	}
	return res
}

func getUserInput() bool {
	var input string
	fmt.Print("refreshing is very dangerous, it will rewrite all the files in the repository, are you sure? (y/n):")
	fmt.Scanln(&input)
	if input == "y" {
		return true
	} else {
		return false
	}
}

func Lock(command KeyCommand) {
	//log.Warning("gitenc lock is not implemented yet.")
	keyPath, _, _ := getKey(command)
	if _, err := os.Stat(keyPath); err != nil {
		log.Error("gitenc isnot initialized in this repository. Run 'gitenc init' to initialize it.")
		return
	}
	encryptFiles := getEncryptFiles()
	ClearGitConfig(command.KeyName)
	for _, file := range encryptFiles {
		RunCommand("git", "checkout", "--", file)
	}

	if getUserInput() {
		RunCommand("git", "rm", "-r", "--cached", "--", getRepoRoot())
		RunCommand("git", "reset", "--hard")
	}
}

func Unlock(command KeyCommand) {
	keyPath, keyName, key := getKey(command)
	if _, err := os.Stat(keyPath); err != nil {
		log.Error("gitenc isnot initialized in this repository. Run 'gitenc init' to initialize it.")
		return
	}
	SetGitConfig(keyName)
	for _, file := range getEncryptFiles() {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Error(err)
			continue
		}

		if len(data) > int(unsafe.Sizeof(HEADER{})) {
			header := (*HEADER)(unsafe.Pointer(&data[0]))
			if header.flag == [4]byte{0, 'M', 'R', 0} && header.version == 1 {
				if header.khash != Hash(key) {
					log.Error("gitenc key is not the same as the one used to encrypt the file.")
					break
				}
				log.Info("Decrypting file: " + file)
				// git add -- filename
				RunCommand("git", "add", "--", file)
				// git checkout -- filename
				RunCommand("git", "checkout", "--", file)
				continue
			}
		}
	}

	if getUserInput() {
		RunCommand("git", "rm", "-r", "--cached", "--", getRepoRoot())
		RunCommand("git", "reset", "--hard")
	}

}

func Doctor(cmd DoctorCommand) {
	// git ls-files -cotsz --exclude-standard ...
	code, output := RunCommand("git", "ls-files", "-cotsz", "--exclude-standard", getRepoRoot())
	if code == 1 {
		log.Error(output)
		return
	}
	fileList := bytes.Split([]byte(output), []byte("\000"))
	encrypted := make([]string, 0)
	unencrypted := make([]string, 0)
	for _, fileInfo := range fileList {
		if len(fileInfo) == 0 {
			continue
		}
		fields := strings.Fields(string(fileInfo))
		if fields[0] == "?" {
			continue
		}
		// git check-attr filter diff -- filename
		code, output := RunCommand("git", "check-attr", "filter", "diff", "--", fields[4])
		if code == 1 {
			log.Error(output)
			return
		}
		attrs := strings.Fields(string(output))
		filter, diff := attrs[2], attrs[5]

		if filter == "gitenc" && diff == "gitenc" {
			encrypted = append(encrypted, fields[4])
			// git cat-file blob object_id
			if blobIsEncrypted(fields[2]) {
				continue
			}

			if cmd.Fix {
				// fix header
				// git add -- filename
				RunCommand("git", "add", "--", fields[4])
				// git ls-files -sz filename
				_, output = RunCommand("git", "ls-files", "-sz", fields[4])
				fields = strings.Fields(output)
				// git cat-file blob object_id
				if blobIsEncrypted(fields[1]) {
					log.Info("Fixed file:", fields[3])
					continue
				}
				log.Error("Failed to fix file:", fields[3])
			} else {
				log.Warning("File is not encrypted:", fields[4], ". Run 'gitenc doctor --fix' to fix it or add it to .gitattributes if you want to ignore it")
			}
		} else {
			unencrypted = append(unencrypted, fields[4])
		}

	}
	if len(encrypted) > 0 {
		log.Info("Encrypted files:")
		for _, file := range encrypted {
			log.Log(file)
		}
	}
	if len(unencrypted) > 0 {
		log.Info("Unencrypted files:")
		for _, file := range unencrypted {
			log.Log(file)
		}
	}
}

func Init(command KeyCommand) {
	//git rev-parse
	code, res := RunCommand("git", "rev-parse")
	if code == 1 {
		log.Error(res)
		return
	}
	keyPath, keyName, key := getKey(command)
	if _, err := os.Stat(keyPath); err == nil {
		log.Error("gitenc is already initialized")
		return
	}
	if err := os.MkdirAll(keyPath, 0700); err != nil {
		log.Error("Error creating key directory", err)
		return
	}
	if err := os.WriteFile(keyPath+keyName, key, 0600); err != nil {
		log.Error("Error writing key", err)
		return
	}
	os.WriteFile(".gitattributes", []byte("* filter=gitenc diff=gitenc\n.gitattributes !filter !diff"), 0600)
	log.Info("gitenc initialized")

	command.KeyName = keyName
	Unlock(command)
}

func getKey(command KeyCommand) (string, string, []byte) {
	keyPath, keyName := GetKeyPath(command.KeyName)
	var key []byte
	if _, err := os.Stat(keyPath + keyName); err != nil {
		if command.Key == "" {
			key := make([]byte, 32)
			rand.Read(key)
			command.Key = string(key)
		}
		key = GenerateKey(command.Key)
	} else {
		key, err = os.ReadFile(keyPath + keyName)
		if err != nil {
			log.Error("Error reading key", err)
			return "", "", nil
		}
	}

	return keyPath, keyName, key
}

func Smudge(cmd KeyCommand) {
	// Read header from stdin
	headerBytes, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		log.Error("Error reading header", err)
		return
	}
	header := (*HEADER)(unsafe.Pointer(&headerBytes[0]))
	// Read encrypted data from stdin
	if header.flag != [4]byte{0, 'M', 'R', 0} || header.version != 1 {
		log.Warning("File is not encrypted. please run 'gitenc doctor' to fix it.")
		os.Stdout.Write(headerBytes)
		return
	}
	input := headerBytes[unsafe.Sizeof(*header):]
	// Decrypt data
	keyPath, keyName := GetKeyPath(cmd.KeyName)
	key, err := os.ReadFile(keyPath + keyName)
	if err != nil {
		log.Warning("File is not encrypted. please run 'gitenc doctor' to fix it.")
		os.Stdout.Write(headerBytes)
		return
	}

	if header.khash != Hash(key) {
		log.Warning("gitenc key is not the same as the one used to encrypt the file.")
		os.Stdout.Write(headerBytes)
		return
	}

	plaintext, err := Decrypt(input, key)
	if err != nil {
		log.Error("Error decrypting", err)
		os.Stdout.Write(headerBytes)
		return
	}

	if header.fhash != Hash(plaintext) {
		log.Error("File hash mismatch")
		os.Stdout.Write(headerBytes)
		return
	}
	// Write decrypted data to stdout
	os.Stdout.Write(plaintext)
}

func Diff(cmd KeyCommand, file string) {
	// Read header from file
	headerBytes, err := ioutil.ReadFile(file)
	if err != nil || len(headerBytes) < int(unsafe.Sizeof(HEADER{})) {
		//log.Error("Error reading file:", err)
		return
	}
	header := (*HEADER)(unsafe.Pointer(&headerBytes[0]))

	if header.flag != [4]byte{0, 'M', 'R', 0} || header.version != 1 {
		os.Stdout.Write(headerBytes)
		return
	}
	// Read encrypted data from file
	dataOffset := int64(unsafe.Sizeof(*header))
	encryptedData, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("Error reading file:", err)
		return
	}
	encryptedData = encryptedData[dataOffset : dataOffset+int64(header.size)]

	// Decrypt data
	keyPath, keyName := GetKeyPath(cmd.KeyName)
	key, err := os.ReadFile(keyPath + keyName)
	if err != nil {
		log.Error("Error reading key", err)
		return
	}

	if header.khash != Hash(key) {
		log.Warning("gitenc key is not the same as the one used to encrypt the file.")
		return
	}

	plaintext, err := Decrypt(encryptedData, key)
	if err != nil {
		log.Error("Error decrypting", err)
		return
	}
	if header.fhash != Hash(plaintext) {
		log.Error("Hash mismatch")
		return
	}

	// Write decrypted data to stdout
	os.Stdout.Write(plaintext)
}

func Clean(cmd KeyCommand) {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	keyPath, keyName := GetKeyPath(cmd.KeyName)
	key, err := os.ReadFile(keyPath + keyName)
	if err != nil {
		log.Error("Error reading key", err)
		return
	}
	encrypted, err := Encrypt(bytes, key)
	if err != nil {
		log.Error("Error encrypting", err)
		return
	}
	header := HEADER{
		flag:    [4]byte{0, 'M', 'R', 0},
		khash:   Hash(key),
		fhash:   Hash(bytes),
		version: 1,
		size:    uint64(len(encrypted)),
	}

	// Write header to stdout
	os.Stdout.Write((*(*[unsafe.Sizeof(header)]byte)(unsafe.Pointer(&header)))[:])
	// Write encrypted data to stdout
	os.Stdout.Write(encrypted)
}

func Set(cmd KeyCommand) {
	keyPath, keyName := GetKeyPath(cmd.KeyName)
	if _, err := os.Stat(keyPath); err != nil {
		log.Error("gitenc isnot initialized in this repository. Run 'gitenc init' to initialize it.")
		return
	}
	if cmd.Key != "" {
		log.Info("Generating new key...")
		key := GenerateKey(cmd.Key)
		if err := os.MkdirAll(keyPath, 0700); err != nil {
			log.Error("Error creating key directory", err)
			return
		}
		if err := os.WriteFile(keyPath+keyName, key, 0600); err != nil {
			log.Error("Error writing key", err)
			return
		}
	}
	log.Info("Setting key name to", keyName)
	SetGitConfig(keyName)
}
