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
	"flag"
	"log"
	"os"
)

type KeyCommand struct {
	Key     string
	KeyName string
}

func init() {

}
func main() {
	key := KeyCommand{}
	KeyCmd := flag.NewFlagSet("init", flag.ExitOnError)
	KeyCmd.StringVar(&key.Key, "key", "", "Key to use for encryption")
	KeyCmd.StringVar(&key.KeyName, "keyname", "", "Name of the key to use for encryption")

	if len(os.Args) < 2 {
		log.Println("Not enough arguments")
		return
	}

	KeyCmd.Parse(os.Args[2:])

	switch os.Args[1] {
	case "init":
		Init(key)
	case "checkout":
		Checkout()
	case "smudge":
		Smudge(key)
	case "clean":
		Clean(key)
	case "diff":
		Diff(key, os.Args[len(os.Args)-1])
	case "version":
		log.Println("gitenc version 0.1")
	default:
		log.Println("Unknown command")
	}

}
