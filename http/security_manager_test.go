package http

import (
	"testing"

	"gotest.tools/assert"
)

func TestFindAuxiliary(t *testing.T) {
	extension := findAuxiliaryExtension("")
	assert.Equal(t, extension, "")

	extension = findAuxiliaryExtension("hello")
	assert.Equal(t, extension, "")

	extension = findAuxiliaryExtension("additionalfwsomething")
	assert.Equal(t, extension, ".bin")

	extension = findAuxiliaryExtension("REMCTRL12345")
	assert.Equal(t, extension, ".tgz")
}

func TestIsAuxiliary(t *testing.T) {
	isAux := isAuxiliary("additionalfwverinfosomethingsomething")
	assert.Equal(t, isAux, true)

	isAux = isAuxiliary("remctrlSOMETHING")
	assert.Equal(t, isAux, true)

	isAux = isAuxiliary("remsomethingctrlSOMETHING")
	assert.Equal(t, isAux, false)

	isAux = isAuxiliary("")
	assert.Equal(t, isAux, false)
}

func TestGetAuxiliaryFirmwares(t *testing.T) {
	auxExtensionString := ""
	auxFirmwareList := getAuxiliaryFirmwares(auxExtensionString)
	assert.Equal(t, len(auxFirmwareList), 0)

	auxExtensionString = "prefix1:ext1"
	auxFirmwareList = getAuxiliaryFirmwares(auxExtensionString)
	assert.Equal(t, len(auxFirmwareList), 1)
	assert.Equal(t, auxFirmwareList[0].Prefix, "prefix1")
	assert.Equal(t, auxFirmwareList[0].Extension, "ext1")

	auxExtensionString = "prefix1:ext1;prefix2:ext2"
	auxFirmwareList = getAuxiliaryFirmwares(auxExtensionString)
	assert.Equal(t, len(auxFirmwareList), 2)
	assert.Equal(t, auxFirmwareList[0].Prefix, "prefix1")
	assert.Equal(t, auxFirmwareList[0].Extension, "ext1")
	assert.Equal(t, auxFirmwareList[1].Prefix, "prefix2")
	assert.Equal(t, auxFirmwareList[1].Extension, "ext2")

	auxExtensionString = "prefix1:ext1;prefix2:ext2;prefix3"
	auxFirmwareList = getAuxiliaryFirmwares(auxExtensionString)
	assert.Equal(t, len(auxFirmwareList), 2)
	assert.Equal(t, auxFirmwareList[0].Prefix, "prefix1")
	assert.Equal(t, auxFirmwareList[0].Extension, "ext1")
	assert.Equal(t, auxFirmwareList[1].Prefix, "prefix2")
	assert.Equal(t, auxFirmwareList[1].Extension, "ext2")

}
