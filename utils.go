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
	"os/exec"
	"strings"
)

func RunCommand(cmd string, args ...string) (int, string) {
	osCmd := exec.Command(cmd, args...)
	output, err := osCmd.CombinedOutput()
	if err != nil {
		return 1, string(output)
	}
	return 0, string(output)
}

func GetKeyPath(name string) (string, string) {
	if name == "" {
		name = "default"
	}
	return GetGitPath() + "/gitenc/keys/", name
}

func Trim(s string) string {
	return strings.Trim(strings.Trim(s, "\r"), "\n")
}
func GetGitPath() string {
	code, path := RunCommand("git", "rev-parse", "--git-dir")
	if code != 0 {
		return ""
	}
	return Trim(path)
}
