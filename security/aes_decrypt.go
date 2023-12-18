/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package security

import (
	"crypto/aes"
	"crypto/cipher"
	b64 "encoding/base64"
	_ "fmt"
)

func Decrypt(xpckeyB64 string, encryptedB64 string) string {
	// CBC decryption

	key, err := b64.StdEncoding.DecodeString(xpckeyB64)
	if err != nil {
		panic(err)
	}
	ciphertext, err := b64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Adapted from Go example code, which is:
	// Copyright 2012 The Go Authors. All rights reserved.
	// Licensed under the BSD-3 License

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point. For an example, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
	// critical to note that ciphertexts must be authenticated (i.e. by
	// using crypto/hmac) before being decrypted in order to avoid creating
	// a padding oracle.

	// unpadding
	index := len(ciphertext) - 1

	for {
		if ciphertext[index] == '\x00' || ciphertext[index] == '\x80' {
			index--
		} else {
			break
		}
	}

	decrypted := ciphertext[20 : index+1]
	return string(decrypted)

	// fmt.Printf("|%v|\n", ciphertext[20:index+1])
	// fmt.Printf("|%s|\n", ciphertext[20:index+1])
}
