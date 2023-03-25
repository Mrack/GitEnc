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
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

func Md5sum(plaintext string) []byte {
	md5 := md5.New()
	md5.Write([]byte(plaintext))
	return md5.Sum(nil)
}
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

func GenerateKey(key string) []byte {
	return append(Md5sum(key), Md5sum(ReverseString(key))...)
}

func Encrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(plainText); err != nil {
		return nil, fmt.Errorf("unable to compress plaintext: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("unable to close gzip writer: %v", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, compressed.Bytes(), nil)
	return ciphertext, nil
}

func Decrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(plainText)
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, fmt.Errorf("error creating gzip reader: %v", err)
	}
	defer r.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, fmt.Errorf("error decompressing data: %v", err)
	}
	return out.Bytes(), nil
}

func Hash(data []byte) [16]byte {
	hash := md5.New()
	hash.Write(data)
	sum := hash.Sum(nil)
	return *(*[16]byte)(unsafe.Pointer(&sum[0]))
}
