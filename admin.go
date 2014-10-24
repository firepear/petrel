// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The server of the everyday task management system

package adminsock

import (
	"log"
)

const (
	VERSION = "0.3.0"
)


func New() {
	log.Printf("Starting everydayd %v\n", VERSION)

	quitter := make(chan bool, 1)     // our master off-switch channel
	admrelaunch := make(chan bool, 1) // our relaunch adm connection notifier

	// launch websocket listener
	go wsHandler()
	log.Println("websocket handler created and launched")

	// launch evdadm listener
	l := launchAdmListener()
	defer l.Close()
	go admAccept(l, quitter, admrelaunch)
	log.Println("evdadm listener created and launched")

	// launch debug logging
	go debugMarktime()

ListenLoop:
	for {
		select {
		case <-admrelaunch:
			l = launchAdmListener()
			go admAccept(l, quitter, admrelaunch)
		case <-quitter:
			log.Println("Got data on quitter channel. Shutting down.")
			break ListenLoop
		}
	}
}
