// Copyright © 2021, 2022 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package walletsharedimport

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	file    []byte
	shares  []string
}

func input(ctx context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	if viper.GetString("remote") != "" {
		return nil, errors.New("wallet import not available for remote wallets")
	}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")

	// Data.
	if viper.GetString("file") == "" {
		return nil, errors.New("file is required")
	}
	data.file, err = os.ReadFile(viper.GetString("file"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read wallet import file")
	}

	// Shares.
	data.shares = viper.GetStringSlice("shares")
	if len(data.shares) == 0 {
		return nil, errors.New("failed to obtain shares")
	}

	return data, nil
}
