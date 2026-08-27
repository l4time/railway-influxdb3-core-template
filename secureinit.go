package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	dataPath  = "/data"
	tokenName = "admin-token.json"
	targetUID = 1500
	targetGID = 1500
)

type tokenDocument struct {
	Token string `json:"token"`
}

type inodeAudit struct {
	TemporaryDescriptorHeldThroughGeneration bool   `json:"temporary_descriptor_held_through_generation"`
	HeldInodeMatchedReservedPathBeforeRename bool   `json:"held_inode_matched_reserved_path_before_rename"`
	FinalInodeMatchedHeldAfterRename          bool   `json:"final_inode_matched_held_after_rename"`
	CaptureTransport                          string `json:"capture_transport"`
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, "fatal:", message)
	os.Exit(70)
}

func fstatType(fd int, want uint32, label string) syscall.Stat_t {
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		fail("inspect " + label + ": " + err.Error())
	}
	if st.Mode&syscall.S_IFMT != want {
		fail(label + " has an unsafe filesystem type")
	}
	return st
}

func openData() int {
	fd, err := syscall.Open(dataPath, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		fail("open data directory without following links: " + err.Error())
	}
	fstatType(fd, syscall.S_IFDIR, "data directory")
	return fd
}

func openToken(dirfd int) (int, bool) {
	fd, err := syscall.Openat(dirfd, tokenName, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if errors.Is(err, syscall.ENOENT) {
		return -1, false
	}
	if err != nil {
		fail("open protected token without following links: " + err.Error())
	}
	fstatType(fd, syscall.S_IFREG, "protected token")
	return fd, true
}

func readAndValidateToken(fd int) []byte {
	duplicate, err := syscall.Dup(fd)
	if err != nil {
		fail("duplicate protected token descriptor: " + err.Error())
	}
	file := os.NewFile(uintptr(duplicate), tokenName)
	defer file.Close()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		fail("seek protected token: " + err.Error())
	}
	raw, err := io.ReadAll(io.LimitReader(file, 1<<20))
	if err != nil {
		fail("read protected token: " + err.Error())
	}
	var doc tokenDocument
	if json.Unmarshal(raw, &doc) != nil || doc.Token == "" {
		fail("protected token is invalid")
	}
	return raw
}

func sameInode(left, right syscall.Stat_t) bool {
	return left.Dev == right.Dev && left.Ino == right.Ino
}

func pathMatchesFD(dirfd int, name string, held syscall.Stat_t) bool {
	fd, err := syscall.Openat(dirfd, name, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return false
	}
	defer syscall.Close(fd)
	var pathStat syscall.Stat_t
	if syscall.Fstat(fd, &pathStat) != nil {
		return false
	}
	return sameInode(held, pathStat)
}

func writeAll(fd int, raw []byte) {
	if _, err := syscall.Seek(fd, 0, io.SeekStart); err != nil {
		fail("seek authoritative protected token descriptor: " + err.Error())
	}
	if err := syscall.Ftruncate(fd, 0); err != nil {
		fail("truncate authoritative protected token descriptor: " + err.Error())
	}
	for len(raw) > 0 {
		written, err := syscall.Write(fd, raw)
		if err != nil {
			fail("write authoritative protected token descriptor: " + err.Error())
		}
		raw = raw[written:]
	}
}

func createToken(dirfd int) (int, inodeAudit) {
	var fd int
	var tempName string
	for attempt := 0; attempt < 32; attempt++ {
		tempName = ".admin-token.r4-" + strconv.Itoa(os.Getpid()) + "-" + strconv.Itoa(attempt)
		var err error
		fd, err = syscall.Openat(
			dirfd,
			tempName,
			syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC,
			0600,
		)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EEXIST) {
			fail("create protected token temporary file: " + err.Error())
		}
	}
	if fd < 0 {
		fail("could not reserve a protected token temporary file")
	}
	defer func() {
		_ = syscall.Unlinkat(dirfd, tempName)
	}()
	if err := syscall.Fchown(fd, targetUID, targetGID); err != nil {
		fail("own protected token temporary file: " + err.Error())
	}
	if err := syscall.Fchmod(fd, 0600); err != nil {
		fail("mode protected token temporary file: " + err.Error())
	}

	var heldStat syscall.Stat_t
	if err := syscall.Fstat(fd, &heldStat); err != nil {
		fail("inspect authoritative protected token descriptor: " + err.Error())
	}

	captureDir, err := os.MkdirTemp("/tmp", ".influx-token-capture-")
	if err != nil {
		fail("create private token capture directory: " + err.Error())
	}
	if err := os.Chmod(captureDir, 0700); err != nil {
		_ = os.RemoveAll(captureDir)
		fail("mode private token capture directory: " + err.Error())
	}
	if err := os.Chown(captureDir, targetUID, targetGID); err != nil {
		_ = os.RemoveAll(captureDir)
		fail("own private token capture directory: " + err.Error())
	}
	capturePath := filepath.Join(captureDir, "offline-token.json")
	cmd := exec.Command(
		"influxdb3", "create", "token", "--admin", "--offline",
		"--output-file", capturePath, "--format", "json",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: targetUID, Gid: targetGID},
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(captureDir)
		fail("offline token generation failed")
	}
	captureFD, err := syscall.Open(capturePath, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		_ = os.RemoveAll(captureDir)
		fail("open private token capture without following links: " + err.Error())
	}
	var captureStat syscall.Stat_t
	if err := syscall.Fstat(captureFD, &captureStat); err != nil || captureStat.Mode&syscall.S_IFMT != syscall.S_IFREG {
		_ = syscall.Close(captureFD)
		_ = os.RemoveAll(captureDir)
		fail("private token capture has an unsafe filesystem type")
	}
	captureFile := os.NewFile(uintptr(captureFD), "private-token-capture")
	raw, readErr := io.ReadAll(io.LimitReader(captureFile, 1<<20))
	_ = captureFile.Close()
	removeErr := os.RemoveAll(captureDir)
	if readErr != nil {
		fail("read private token capture")
	}
	if removeErr != nil {
		fail("remove private token capture")
	}
	var captured tokenDocument
	if json.Unmarshal(raw, &captured) != nil || captured.Token == "" {
		fail("private token capture is invalid")
	}

	writeAll(fd, raw)
	if err := syscall.Fsync(fd); err != nil {
		fail("sync authoritative protected token descriptor: " + err.Error())
	}
	readAndValidateToken(fd)

	if _, exists := openToken(dirfd); exists {
		fail("protected token appeared during initialization")
	}
	if !pathMatchesFD(dirfd, tempName, heldStat) {
		fail("authoritative protected token entry changed before install")
	}
	if err := syscall.Renameat(dirfd, tempName, dirfd, tokenName); err != nil {
		fail("install protected token: " + err.Error())
	}
	finalFD, err := syscall.Openat(dirfd, tokenName, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		fail("reopen installed token without following links: " + err.Error())
	}
	var finalStat syscall.Stat_t
	if err := syscall.Fstat(finalFD, &finalStat); err != nil {
		fail("inspect installed token: " + err.Error())
	}
	_ = syscall.Close(finalFD)
	if !sameInode(heldStat, finalStat) {
		fail("installed token does not match authoritative descriptor")
	}
	if err := syscall.Fsync(dirfd); err != nil {
		fail("sync data directory: " + err.Error())
	}
	return fd, inodeAudit{
		TemporaryDescriptorHeldThroughGeneration: true,
		HeldInodeMatchedReservedPathBeforeRename: true,
		FinalInodeMatchedHeldAfterRename:          true,
		CaptureTransport:                          "private-mode-0700-capture-copied-to-held-descriptor",
	}
}

func runInodeAuthoritySelfTest() {
	scratch, err := os.MkdirTemp("/tmp", ".secure-init-inode-self-test-")
	if err != nil {
		fail("create inode authority self-test directory")
	}
	defer os.RemoveAll(scratch)
	dirfd, err := syscall.Open(scratch, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		fail("open inode authority self-test directory")
	}
	defer syscall.Close(dirfd)
	held, err := syscall.Openat(dirfd, "candidate", syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		fail("create inode authority self-test candidate")
	}
	defer syscall.Close(held)
	if _, err := syscall.Write(held, []byte("held-inode")); err != nil {
		fail("write inode authority self-test candidate")
	}
	var heldStat syscall.Stat_t
	if syscall.Fstat(held, &heldStat) != nil {
		fail("inspect inode authority self-test candidate")
	}
	alternate, err := syscall.Openat(dirfd, "alternate", syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		fail("create inode authority self-test alternate")
	}
	_, _ = syscall.Write(alternate, []byte("alternate-inode"))
	_ = syscall.Close(alternate)
	if syscall.Unlinkat(dirfd, "candidate") != nil || syscall.Renameat(dirfd, "alternate", dirfd, "candidate") != nil {
		fail("prepare inode authority mismatch self-test")
	}
	if pathMatchesFD(dirfd, "candidate", heldStat) {
		fail("inode authority self-test accepted a different entry")
	}
	if _, err := syscall.Seek(held, 0, io.SeekStart); err != nil {
		fail("seek inode authority self-test held descriptor")
	}
	heldContent := make([]byte, len("held-inode"))
	read, err := syscall.Read(held, heldContent)
	if err != nil || read != len(heldContent) || string(heldContent) != "held-inode" {
		fail("inode authority self-test held content changed")
	}
	fmt.Println(`{"alternate_entry_rejected":true,"held_descriptor_content_unchanged":true,"safe_fixture_only":true}`)
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--self-test-inode-authority" {
		runInodeAuthoritySelfTest()
		return
	}
	if len(os.Args) != 1 {
		fail("unexpected argument")
	}
	if os.Geteuid() != 0 {
		fail("secure init must start as root")
	}
	dirfd := openData()
	defer syscall.Close(dirfd)

	// Validate the final token entry before *any* root metadata operation.
	tokenfd, exists := openToken(dirfd)
	if exists {
		defer syscall.Close(tokenfd)
	}

	if err := syscall.Fchown(dirfd, targetUID, targetGID); err != nil {
		fail("own data directory: " + err.Error())
	}
	if err := syscall.Fchmod(dirfd, 0750); err != nil {
		fail("mode data directory: " + err.Error())
	}
	if !exists {
		var audit inodeAudit
		tokenfd, audit = createToken(dirfd)
		defer syscall.Close(tokenfd)
		raw, err := json.Marshal(audit)
		if err != nil {
			fail("encode secure init audit")
		}
		if err := os.WriteFile("/run/secure-init-audit.json", append(raw, '\n'), 0444); err != nil {
			fail("write secure init audit")
		}
	}
	fstatType(tokenfd, syscall.S_IFREG, "protected token")
	if err := syscall.Fchown(tokenfd, targetUID, targetGID); err != nil {
		fail("own protected token: " + err.Error())
	}
	if err := syscall.Fchmod(tokenfd, 0600); err != nil {
		fail("mode protected token: " + err.Error())
	}
	readAndValidateToken(tokenfd)
}
