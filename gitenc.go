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
	log "gitenc/log"
	"os"
)

type KeyCommand struct {
	Key     string
	KeyName string
}

type DoctorCommand struct {
	Fix bool
}

func init() {
	log.InitLog()
}

func main() {
	key := KeyCommand{}
	KeyCmd := flag.NewFlagSet("init", flag.ExitOnError)
	KeyCmd.StringVar(&key.Key, "key", "", "Key to use for encryption")
	KeyCmd.StringVar(&key.KeyName, "keyname", "", "Name of the key to use for encryption")

	doctor := DoctorCommand{}
	DoctorCmd := flag.NewFlagSet("doctor", flag.ExitOnError)
	DoctorCmd.BoolVar(&doctor.Fix, "fix", false, "Fix problems")

	if len(os.Args) < 2 {
		log.Error("Not enough arguments")
		showHelp()
		return
	}

	//log.Info(os.Args)

	switch os.Args[1] {
	case "init":
		KeyCmd.Parse(os.Args[2:])
		Init(key)
	case "set":
		KeyCmd.Parse(os.Args[2:])
		Set(key)
	case "lock":
		KeyCmd.Parse(os.Args[2:])
		Lock(key)
	case "unlock":
		KeyCmd.Parse(os.Args[2:])
		Unlock(key)
	case "doctor":
		DoctorCmd.Parse(os.Args[2:])
		Doctor(doctor)
	case "smudge":
		KeyCmd.Parse(os.Args[2:])
		Smudge(key)
	case "clean":
		KeyCmd.Parse(os.Args[2:])
		Clean(key)
	case "diff":
		KeyCmd.Parse(os.Args[2:])
		Diff(key, os.Args[len(os.Args)-1])
	case "version":
		log.Info("gitenc version 0.4")
	case "help":
		showHelp()
	default:
		log.Warning("Unknown command: " + os.Args[1] + ". Try 'gitenc help' for more information.")
	}

}

func showHelp() {
	log.Info("Usage: gitenc <command> [options]")
	log.Info("Commands:")
	log.Log("init - Initialize gitenc in the current repository")
	log.Log("set - Set the key to use for encryption")
	log.Log("lock - Lock the repository")
	log.Log("unlock - Unlock the repository")
	log.Log("doctor - Check the repository for problems")
	log.Log("version - Print the version of gitenc")
	log.Log("help - Print this help message")
}
