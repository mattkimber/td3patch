package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	slowdown := flag.Int("delay", 6, "The number of frames to wait (1-255)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: accopatch -delay [num] [path_to_exe]")
		os.Exit(1)
	}
	exePath := args[0]

	if *slowdown < 1 || *slowdown > 255 {
		fmt.Println("Error: Delay must be between 1 and 255")
		os.Exit(1)
	}

	delayByte := byte(*slowdown)

	err := patchExecutable(exePath, delayByte, slowdown)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func patchExecutable(exePath string, delayByte byte, slowdown *int) error {
	data, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("could not read file: %v\n", err)
	}

	foundIdx, doBackup, err := findVBlankPattern(exePath, delayByte, slowdown, data)
	if err != nil {
		return err
	}

	// Target: 51 B1 xx BA DA 03 EC A8 08 74 FB EC A8 08 75 FB FE C9 75 F2 59 90
	newSequence := []byte{
		0x51, 0xB1, delayByte, 0xBA, 0xDA, 0x03, 0xEC, 0xA8, 0x08, 0x74,
		0xFB, 0xEC, 0xA8, 0x08, 0x75, 0xFB, 0xFE, 0xC9, 0x75, 0xF2, 0x59, 0x90,
	}

	if doBackup {
		backupPath := exePath + ".BAK"
		if _, err := os.Stat(backupPath); err == nil {
			fmt.Println(".BAK file already exists. Skipping backup.")
		} else if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("failure creating backup: %v\n", err)
		}
	}

	// Save new file
	fmt.Printf("Updating pattern at offset 0x%X to apply new %d-frame delay patch.\n", foundIdx, *slowdown)
	for k := 0; k < len(newSequence); k++ {
		data[foundIdx+k] = newSequence[k]
	}
	return saveFile(exePath, data)
}

func findVBlankPattern(exePath string, delayByte byte, slowdown *int, data []byte) (int, bool, error) {
	// Pattern Definitions
	// Original: BA DA 03 EC A8 08 74 F8 80 3E __ __ 00 75 07 80 3E __ __ 00 74 06
	searchPattern := []byte{
		0xBA, 0xDA, 0x03, 0xEC, 0xA8, 0x08, 0x74, 0xF8, 0x80, 0x3E,
		0x00, 0x00, // Memory var changes per EXE
		0x00, 0x75, 0x07, 0x80, 0x3E,
		0x00, 0x00, // Memory var changes per EXE
		0x00, 0x74, 0x06,
	}

	for i := 0; i <= len(data)-len(searchPattern); i++ {

		// Check if already patched: Starts with 51 B1 and matches BA DA 03 at index i+3
		if data[i] == 0x51 && data[i+1] == 0xB1 &&
			data[i+3] == 0xBA && data[i+4] == 0xDA && data[i+5] == 0x03 {

			fmt.Printf("Existing patch found at offset 0x%X.\n", i)
			return i, false, nil
		}

		// Check for original pattern with wildcard masks
		match := true
		for j := 0; j < len(searchPattern); j++ {
			// Skip variable memory addresses (Indices 10, 11, 17, 18)
			if j == 10 || j == 11 || j == 17 || j == 18 {
				continue
			}
			if data[i+j] != searchPattern[j] {
				match = false
				break
			}
		}

		if match {
			fmt.Printf("Original code found at offset 0x%X.\n", i)
			return i, true, nil
		}
	}

	return -1, false, errors.New("could not find the target byte sequence in this file")
}

func saveFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %v\n", err)
	}
	fmt.Println("Success!")
	return nil
}
