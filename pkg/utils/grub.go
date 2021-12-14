package utils

import (
	"fmt"
	"github.com/rancher-sandbox/elemental-cli/pkg/constants"
	v1 "github.com/rancher-sandbox/elemental-cli/pkg/types/v1"
	"github.com/spf13/afero"
	"runtime"
	"strings"
)

type Grub struct {
	disk   string
	config *v1.RunConfig
}

func NewGrub(config *v1.RunConfig) *Grub {
	g := &Grub{
		config: config,
	}

	return g
}

func (g Grub) Install() error {
	var grubargs []string
	var arch, grubdir, tty, finalContent string
	var err error

	switch runtime.GOARCH {
	case "arm64":
		arch = "arm64"
	default:
		arch = "x86_64"
	}
	g.config.Logger.Info("Installing GRUB..")

	if g.config.Tty == "" {
		// Get current tty and remove /dev/ from its name
		out, err := g.config.Runner.Run("tty")
		tty = strings.TrimPrefix(strings.TrimSpace(string(out)), "/dev/")
		if err != nil {
			return err
		}
	} else {
		tty = g.config.Tty
	}

	// check if dir exists before creating them
	for _, d := range []string{"proc", "dev", "sys", "tmp"} {
		createDir := fmt.Sprintf("%s/%s", g.config.Target, d)
		if exists, _ := afero.DirExists(g.config.Fs, createDir); !exists {
			err = g.config.Fs.Mkdir(createDir, 0644)
			if err != nil {
				return err
			}
		}
	}

	efiExists, _ := afero.Exists(g.config.Fs, constants.EfiDevice)

	if g.config.ForceEfi || efiExists {
		g.config.Logger.Infof("Installing grub efi for arch %s", arch)
		grubargs = append(
			grubargs,
			fmt.Sprintf("--target=%s-efi", arch),
			fmt.Sprintf("--efi-directory=%s/boot/efi", g.config.Target),
		)
	}

	grubargs = append(
		grubargs,
		fmt.Sprintf("--root-directory=%s", g.config.StateDir),
		fmt.Sprintf("--boot-directory=%s", g.config.StateDir),
		fmt.Sprintf("--removable=%s", g.config.Device),
	)

	g.config.Logger.Debugf("Running grub with the following args: %s", grubargs)
	_, err = g.config.Runner.Run("grub2-install", grubargs...)
	if err != nil {
		return err
	}

	grub1dir := fmt.Sprintf("%s/grub", g.config.StateDir)
	grub2dir := fmt.Sprintf("%s/grub2", g.config.StateDir)

	// Select the proper dir for grub
	if ok, _ := afero.IsDir(g.config.Fs, grub1dir); ok {
		grubdir = grub1dir
	}
	if ok, _ := afero.IsDir(g.config.Fs, grub2dir); ok {
		grubdir = grub2dir
	}
	g.config.Logger.Infof("Found grub config dir %s", grubdir)

	grubConf, err := afero.ReadFile(g.config.Fs, g.config.GrubConf)

	grubConfTarget, err := g.config.Fs.Create(fmt.Sprintf("%s/grub.cfg", grubdir))
	defer grubConfTarget.Close()

	ttyExists, _ := afero.Exists(g.config.Fs, fmt.Sprintf("/dev/%s", tty))

	if ttyExists && tty != "" && tty != "console" && tty != "tty1" {
		// We need to add a tty to the grub file
		g.config.Logger.Infof("Adding extra tty (%s) to grub.cfg", tty)
		finalContent = strings.Replace(string(grubConf), "console=tty1", fmt.Sprintf("console=tty1 console=%s", tty), -1)
	} else {
		// We don't add anything, just read the file
		finalContent = string(grubConf)
	}

	g.config.Logger.Infof("Copying grub contents from %s to %s", g.config.GrubConf, fmt.Sprintf("%s/grub.cfg", grubdir))
	_, err = grubConfTarget.WriteString(finalContent)
	if err != nil {
		return err
	}

	g.config.Logger.Infof("Grub install to device %s complete", g.config.Device)
	return nil
}
