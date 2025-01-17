/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lock

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/juju/fslock"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/util/retry"
)

// WriteWithLock decorates ioutil.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) (err error) {
	lock := fslock.New(filename)
	glog.Infof("attempting to write to file %q with filemode %v", filename, perm)

	getLock := func() error {
		lockErr := lock.TryLock()
		if lockErr != nil {
			glog.Warningf("temporary error : %v", lockErr.Error())
			return errors.Wrapf(lockErr, "failed to acquire lock for %s > ", filename)
		}
		return nil
	}

	defer func() { // release the lock
		err = lock.Unlock()
		if err != nil {
			err = errors.Wrapf(err, "error releasing lock for file: %s", filename)
		}
	}()

	err = retry.Expo(getLock, 500*time.Millisecond, 13*time.Second)
	if err != nil {
		return errors.Wrapf(err, "error acquiring lock for %s", filename)
	}

	if err = ioutil.WriteFile(filename, data, perm); err != nil {
		return errors.Wrapf(err, "error writing file %s", filename)
	}

	return err
}
