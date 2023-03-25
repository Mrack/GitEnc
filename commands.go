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
	"crypto/rand"
	"io/ioutil"
	"log"
	"os"
	"unsafe"
)

type HEADER struct {
	flag    [4]byte
	version byte
	khash   [16]byte
	fhash   [16]byte
	size    uint64
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

func Checkout() {
	// git checkout-index -f --all
	RunCommand("git", "checkout")
}

func Init(command KeyCommand) {
	keyPath, keyName := GetKeyPath(command.KeyName)
	if _, err := os.Stat(keyPath); err == nil {
		log.Println("gitenc already initialized")
		return
	}
	if command.Key == "" {
		key := make([]byte, 32)
		rand.Read(key)
		command.Key = string(key)
	}
	key := GenerateKey(command.Key)
	if err := os.MkdirAll(keyPath, 0700); err != nil {
		log.Println("Error creating key directory", err)
		return
	}
	if err := os.WriteFile(keyPath+keyName, key, 0600); err != nil {
		log.Println("Error writing key", err)
		return
	}
	SetGitConfig(keyName)

	os.WriteFile(".gitattributes", []byte("* filter=gitenc diff=gitenc\n.gitattributes !filter !diff"), 0600)
	log.Println("gitenc initialized")
}

func Smudge(cmd KeyCommand) {
	// Read header from stdin
	headerBytes, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		log.Println("Error reading header", err)
		return
	}
	header := (*HEADER)(unsafe.Pointer(&headerBytes[0]))
	// Read encrypted data from stdin
	if header.flag != [4]byte{0, 'M', 'R', 0} || header.version != 1 {
		os.Stdout.Write(headerBytes)
		return
	}
	input := headerBytes[unsafe.Sizeof(*header):]
	// Decrypt data
	keyPath, keyName := GetKeyPath(cmd.KeyName)
	key, err := os.ReadFile(keyPath + keyName)
	if err != nil {
		log.Println("Error reading key", err)
		os.Stdout.Write(headerBytes)
		return
	}

	if header.khash != Hash(key) {
		log.Println("Key mismatch")
		os.Stdout.Write(headerBytes)
		return
	}

	plaintext, err := Decrypt(input, key)
	if err != nil {
		log.Println("Error decrypting", err)
		os.Stdout.Write(headerBytes)
		return
	}

	if header.fhash != Hash(plaintext) {
		log.Println("File hash mismatch")
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
		log.Println("Error reading file:", err)
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
		log.Println("Error reading file:", err)
		return
	}
	encryptedData = encryptedData[dataOffset : dataOffset+int64(header.size)]

	// Decrypt data
	keyPath, keyName := GetKeyPath(cmd.KeyName)
	key, err := os.ReadFile(keyPath + keyName)
	if err != nil {
		log.Println("Error reading key", err)
		return
	}

	if header.khash != Hash(key) {
		log.Println("Key mismatch")
		return
	}

	plaintext, err := Decrypt(encryptedData, key)
	if err != nil {
		log.Println("Error decrypting", err)
		return
	}
	if header.fhash != Hash(plaintext) {
		log.Println("Hash mismatch")
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
		log.Println("Error reading key", err)
		return
	}
	encrypted, err := Encrypt(bytes, key)
	if err != nil {
		log.Println("Error encrypting", err)
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
