package container

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// POST "/containers/create"
func (cr *containerRouter) ContainerCreate(c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	ctainr, err := cr.factory.Create()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	ctainr.SetName(in.Name)
	ctainr.SetImage(in.Image)
	ctainr.SetCmd(in.Cmd)
	ctainr.SetEnv(in.Env)
	ctainr.SetExposedPorts(in.ExposedPorts)
	ctainr.SetLabels(in.Labels)
	ctainr.Update()

	c.JSON(http.StatusCreated, gin.H{
		"Id": ctainr.GetID(),
	})
}

// POST "/containers/:id/start"
func (cr *containerRouter) ContainerStart(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := cr.kubernetes.StartContainer(ctainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// DELETE "/containers/:id"
func (cr *containerRouter) ContainerDelete(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := cr.kubernetes.DeleteContainer(ctainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	if err := ctainr.Delete(); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// GET "/containers/:id/json"
func (cr *containerRouter) ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	status, err := cr.kubernetes.GetContainerStatus(ctainr)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Id":    id,
		"Image": ctainr.GetImage(),
		"Config": gin.H{
			"Image":  ctainr.GetImage(),
			"Labels": ctainr.GetLabels(),
			"Env":    ctainr.GetEnv(),
			"Cmd":    ctainr.GetCmd(),
		},
		"NetworkSettings": gin.H{
			"Ports": cr.getNetworkSettingsPorts(ctainr),
		},
		"HostConfig": gin.H{
			"NetworkMode": "host",
		},
		"State": gin.H{
			"Health": gin.H{
				"Status": status["Status"],
			},
			"Running":    status["Running"] == "running",
			"Status":     status["Running"],
			"Paused":     false,
			"Restarting": false,
			"OOMKilled":  false,
			"Dead":       false,
			"StartedAt":  "2021-01-01T00:00:00Z",
			"FinishedAt": "0001-01-01T00:00:00Z",
			"ExitCode":   0,
			"Error":      "",
		},
	})
}

// getNetworkSettingsPorts will return the mapped ports of the container
// as k8s ports structure to be used in network settings.
func (cr *containerRouter) getNetworkSettingsPorts(tainr container.Container) gin.H {
	res := gin.H{}
	for dst, src := range tainr.GetMappedPorts() {
		p := fmt.Sprintf("%d/tcp", dst)
		res[p] = []gin.H{
			{
				"HostIp":   "localhost",
				"HostPort": fmt.Sprintf("%d", src),
			},
		}
	}
	return res
}