package backend

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/ioproxy"
)

// GetLogs will write the logs for given container to given writer.
func (in *instance) GetLogs(tainr *types.Container, follow bool, count int, stop chan struct{}, w io.Writer) error {
	tail := int64(count)
	options := v1.PodLogOptions{
		Container: "main",
		Follow:    follow,
		TailLines: &tail,
	}

	name, err := in.getFirstPodName(tainr)
	if err != nil {
		return err
	}

	req := in.cli.CoreV1().Pods(in.namespace).GetLogs(name, &options)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		return err
	}
	defer stream.Close()

	stopL := make(chan struct{}, 1)

	if follow {
		go func() {
			<-stop
			stopL <- struct{}{}
			stream.Close()
		}()
	}

	out := ioproxy.New(w, ioproxy.Stdout)
	defer out.Flush()
	for {
		// close when container is done
		select {
		case <-stopL:
			close(stopL)
			return nil
		default:
		}
		// read log input (blocking read)
		buf := make([]byte, 255)
		n, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if n == 0 {
			if !follow {
				break
			}
			continue
		}
		// write log to output
		if n, err = out.Write(buf[:n]); n == 0 || err != nil {
			break
		}
	}

	return nil
}

// getFirstPodName returns the pod name of the first pod that matches
// the container deployment.
func (in *instance) getFirstPodName(tainr *types.Container) (string, error) {
	pods, err := in.getPods(tainr)
	if err != nil {
		return "", err
	}

	names := []string{}
	for _, p := range pods {
		names = append(names, p.Name)
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no running pods for %s", tainr.ShortID)
	}

	return names[0], nil
}
