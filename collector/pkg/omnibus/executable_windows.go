//go:build windows

package omnibus

func executableName(name string) string {
	return name + ".exe"
}
