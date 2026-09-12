package bootstrap

import (
	"path"
	"time"

	"reasonix/internal/remote/serveenv"
	"reasonix/internal/store"
)

// ServeState is the JSON record a bootstrapped serve leaves on the remote host
// so a later (re)connect can find and reuse it. It is defined in the ssh-free
// serveenv leaf package so non-remote CLI paths can use it without pulling in
// the ssh client stack; this alias preserves the bootstrap.ServeState name.
type ServeState = serveenv.ServeState

// MarshalState renders a ServeState as indented JSON.
func MarshalState(s ServeState) ([]byte, error) { return serveenv.MarshalState(s) }

// UnmarshalState parses a ServeState record.
func UnmarshalState(data []byte) (ServeState, error) { return serveenv.UnmarshalState(data) }

// remoteDir is the ~/.reasonix/remote directory given the resolved remote home.
func remoteDir(home string) string {
	return path.Join(home, ".reasonix", store.RemoteDirName)
}

// pathsFor derives every per-workspace state path from the resolved remote
// home and workspace directory.
func pathsFor(home, workspace string) StatePaths {
	dir := remoteDir(home)
	slug := store.RemoteWorkspaceSlug(workspace)
	return StatePaths{
		Dir:       dir,
		StateJSON: path.Join(dir, store.RemoteServeStateName(slug)),
		TokenFile: path.Join(dir, store.RemoteServeTokenName(slug)),
		LogFile:   path.Join(dir, store.RemoteServeLogName(slug)),
		PortFile:  path.Join(dir, store.RemoteServePortName(slug)),
		PidFile:   path.Join(dir, store.RemoteServePidName(slug)),
		LockDir:   path.Join(dir, store.RemoteServeLockName(slug)),
		LockOwner: path.Join(dir, store.RemoteServeLockName(slug), "owner"),
	}
}

// uploadedBinPath is the fallback location for an uploaded reasonix binary.
func uploadedBinPath(home string) string {
	return path.Join(remoteDir(home), store.RemoteBinDirName, "reasonix")
}

func nowUnix(clock func() time.Time) int64 {
	if clock == nil {
		return 0
	}
	return clock().Unix()
}
