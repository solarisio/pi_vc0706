package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/wujiang/pi_vc0706"
)

func main() {
	s, err := vc0706.InitCamera()
	if err != nil {
		return
	}
	glog.Info("Initialized")
	version, err := vc0706.GetVersion(s)
	fmt.Println(version)

	buf, err := vc0706.TakePhoto(s)
	if err != nil {
		glog.Warning(err)
	}
	glog.Info("Take photo")
	if err = vc0706.SaveBuffer("test.jpg", buf); err != nil {
		glog.Warning(err)
	}
}
