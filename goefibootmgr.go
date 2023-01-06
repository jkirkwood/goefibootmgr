package goefibootmgr

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// BootEntry represents and EFI boot number entry
type BootEntry struct {
	Num    uint16
	Active bool
	Label  string
}

// Activate activates the given boot entry
func (b *BootEntry) Activate() error {
	err := exec.Command("efibootmgr", "-b", bootnumToHexString(b.Num), "-a").Run()

	if err != nil {
		b.Active = true
	}

	return err
}

// Deactivate deactivates the given boot entry
func (b *BootEntry) Deactivate() error {
	err := exec.Command("efibootmgr", "-b", bootnumToHexString(b.Num), "-A").Run()

	if err != nil {
		b.Active = false
	}

	return err
}

// Delete delete the given boot entry
func (b *BootEntry) Delete() error {
	return exec.Command("efibootmgr", "-b", bootnumToHexString(b.Num), "-B").Run()
}

// BootManagerInfo holds all info returned from the efibootmgr command
type BootManagerInfo struct {
	// BootEntry stored in the EFI BootCurrent variable
	BootCurrent *BootEntry
	// BootEntry stored in the EFI BootNext variable
	BootNext *BootEntry
	// Slice containing BootEntrys in order they will boot
	BootOrder []BootEntry
	// Slice containing all detected boot entries
	BootEntries []BootEntry
}

// BootInfo runs the efibootmanager command and returns an BootInfo struct
// containing all the info that was returned
func BootInfo() (*BootManagerInfo, error) {
	out, err := exec.Command("efibootmgr").Output()

	if err != nil {
		return nil, err
	}

	bootCurrentRe := regexp.MustCompile(`BootCurrent: ([0-9a-fA-F]{4})`)
	bootNextRe := regexp.MustCompile(`BootNext: ([0-9a-fA-F]{4})`)
	bootOrderRe := regexp.MustCompile(`BootOrder: ([0-9a-fA-F]{4}(?:,[0-9a-fA-F]{4})*)`)
	bootEntryRe := regexp.MustCompile(`Boot([0-9a-fA-F]{4})(\*?)\s+(.*)`)

	lines := strings.Split(string(out), "\n")

	var bootEntryMap = map[uint16]BootEntry{}

	bm := BootManagerInfo{}

	// First find boot entries
	for _, line := range lines {
		if match := bootEntryRe.FindStringSubmatch(line); match != nil {
			e := BootEntry{
				Num:    hexStringToBootNum(match[1]),
				Active: match[2] == "*",
				Label:  match[3],
			}
			bootEntryMap[e.Num] = e
			bm.BootEntries = append(bm.BootEntries, e)
		}
	}

	// Now parse the rest of the boot info
	for _, line := range lines {
		if match := bootCurrentRe.FindStringSubmatch(line); match != nil {
			num := hexStringToBootNum(match[1])
			entry := bootEntryMap[num]
			bm.BootCurrent = &entry
		} else if match := bootNextRe.FindStringSubmatch(line); match != nil {
			num := hexStringToBootNum(match[1])
			entry := bootEntryMap[num]
			bm.BootNext = &entry
		} else if match := bootOrderRe.FindStringSubmatch(line); match != nil {
			numList := match[1]
			numSlice := strings.Split(numList, ",")
			for _, num := range numSlice {
				num := hexStringToBootNum(num)
				bm.BootOrder = append(bm.BootOrder, bootEntryMap[num])
			}
		}
	}

	return &bm, nil
}

// SetBootOrder sets the EFI BootOrder variable using a list of args containing
// boot numbers
func SetBootOrder(bo ...uint16) (err error) {
	if len(bo) == 0 {
		// Delete BootOrder
		err = exec.Command("efibootmgr", "-O").Run()
	} else {
		s := bootnumToHexString(bo[0])
		for _, o := range bo[1:] {
			s += "," + bootnumToHexString(o)
		}

		err = exec.Command("efibootmgr", "-o", s).Run()
	}

	return
}

// SetBootNext sets the EFI BootNext variable to the bootnum passed to it
func SetBootNext(b uint16) (err error) {
	return exec.Command("efibootmgr", "-n", bootnumToHexString(b)).Run()
}

// DeleteBootNext deletes the EFI BootNext variable
func DeleteBootNext() error {
	return exec.Command("efibootmgr", "-N").Run()
}

func bootnumToHexString(bootnum uint16) string {
	return fmt.Sprintf("%04X", bootnum)
}

func hexStringToBootNum(s string) uint16 {
	b, _ := hex.DecodeString(s)
	return binary.BigEndian.Uint16(b)
}
