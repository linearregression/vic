// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vmware/vic/lib/portlayer/event"
	"github.com/vmware/vic/lib/portlayer/event/collector/vsphere"
	"github.com/vmware/vic/lib/portlayer/event/events"
)

var containerEvents []events.Event

func TestEventedState(t *testing.T) {
	// poweredOn event
	event := events.ContainerPoweredOn
	assert.EqualValues(t, StateStarting, eventedState(event, StateStarting))
	assert.EqualValues(t, StateRunning, eventedState(event, StateRunning))
	assert.EqualValues(t, StateRunning, eventedState(event, StateStopped))
	assert.EqualValues(t, StateRunning, eventedState(event, StateSuspended))

	// powerOff event
	event = events.ContainerPoweredOff
	assert.EqualValues(t, StateStopping, eventedState(event, StateStopping))
	assert.EqualValues(t, StateStopped, eventedState(event, StateStopped))
	assert.EqualValues(t, StateStopped, eventedState(event, StateRunning))

	// suspended event
	event = events.ContainerSuspended
	assert.EqualValues(t, StateSuspending, eventedState(event, StateSuspending))
	assert.EqualValues(t, StateSuspended, eventedState(event, StateSuspended))
	assert.EqualValues(t, StateSuspended, eventedState(event, StateRunning))

	// removed event
	event = events.ContainerRemoved
	assert.EqualValues(t, StateRemoved, eventedState(event, StateRunning))
	assert.EqualValues(t, StateRemoved, eventedState(event, StateStopped))
	assert.EqualValues(t, StateRemoving, eventedState(event, StateRemoving))
}

func TestPublishContainerEvent(t *testing.T) {

	NewContainerCache()
	containerEvents = make([]events.Event, 0)
	VCHConfig = Configuration{}

	mgr := event.NewEventManager()
	VCHConfig.EventManager = mgr
	mgr.Subscribe(events.NewEventType(events.ContainerEvent{}).Topic(), "testing", containerCallback)
	mgr.Subscribe(events.NewEventType(vsphere.VmEvent{}).Topic(), "infra", eventCallback)

	// create new running container and place in cache
	id := "123439"
	container := newTestContainer(id)
	addTestVM(container)
	container.State = StateRunning
	containers.Put(container)

	// create vm PoweredOff event and publish
	ve := &vsphere.VmEvent{
		&events.BaseEvent{
			Event: events.ContainerPoweredOff,
			Ref:   container.vm.Reference().String(),
		},
	}
	mgr.Publish(ve)
	time.Sleep(time.Millisecond * 30)

	assert.Equal(t, 1, len(containerEvents))
	assert.Equal(t, id, containerEvents[0].Reference())
	assert.Equal(t, events.ContainerPoweredOff, containerEvents[0].String())
	assert.EqualValues(t, StateStopped, containers.Container(id).State)

}

func containerCallback(ee events.Event) {
	containerEvents = append(containerEvents, ee)
}
