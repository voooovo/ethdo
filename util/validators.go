// Copyright © 2022 Weald Technology Trading
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

package util

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

// ParseValidators parses input to obtain the list of validators.
func ParseValidators(ctx context.Context, validatorsProvider eth2client.ValidatorsProvider, validatorsStr []string, stateID string) ([]*apiv1.Validator, error) {
	validators := make([]*apiv1.Validator, 0, len(validatorsStr))
	for i := range validatorsStr {
		if strings.Contains(validatorsStr[i], "-") {
			// Range.
			bits := strings.Split(validatorsStr[i], "-")
			if len(bits) != 2 {
				return nil, fmt.Errorf("invalid range %s", validatorsStr[i])
			}
			low, err := strconv.ParseUint(bits[0], 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "invalid range start")
			}
			high, err := strconv.ParseUint(bits[1], 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "invalid range end")
			}
			indices := make([]phase0.ValidatorIndex, 0)
			for index := low; index <= high; index++ {
				indices = append(indices, phase0.ValidatorIndex(index))
			}
			rangeValidators, err := validatorsProvider.Validators(ctx, stateID, indices)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to obtain validators %s", validatorsStr[i]))
			}
			for _, validator := range rangeValidators {
				validators = append(validators, validator)
			}
		} else {
			validator, err := ParseValidator(ctx, validatorsProvider, validatorsStr[i], stateID)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("unknown validator %s", validatorsStr[i]))
			}
			validators = append(validators, validator)
		}
	}
	return validators, nil
}

// ParseValidator parses input to obtain the validator.
func ParseValidator(ctx context.Context,
	validatorsProvider eth2client.ValidatorsProvider,
	validatorStr string,
	stateID string,
) (
	*apiv1.Validator,
	error,
) {
	var validators map[phase0.ValidatorIndex]*apiv1.Validator

	switch {
	case strings.HasPrefix(validatorStr, "0x"):
		// A public key.
		data, err := hex.DecodeString(strings.TrimPrefix(validatorStr, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse validator public key")
		}
		pubKey := phase0.BLSPubKey{}
		copy(pubKey[:], data)
		validators, err = validatorsProvider.ValidatorsByPubKey(ctx,
			stateID,
			[]phase0.BLSPubKey{pubKey},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain validator information")
		}
	case strings.Contains(validatorStr, "/"):
		// An account.
		_, account, err := WalletAndAccountFromPath(ctx, validatorStr)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain account")
		}
		accPubKey, err := BestPublicKey(account)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain public key for account")
		}
		pubKey := phase0.BLSPubKey{}
		copy(pubKey[:], accPubKey.Marshal())
		validators, err = validatorsProvider.ValidatorsByPubKey(ctx,
			stateID,
			[]phase0.BLSPubKey{pubKey},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain validator information")
		}
	default:
		// An index.
		index, err := strconv.ParseUint(validatorStr, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse validator index")
		}
		validators, err = validatorsProvider.Validators(ctx, stateID, []phase0.ValidatorIndex{phase0.ValidatorIndex(index)})
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain validator information")
		}
	}
	// Validator is first entry in the map.
	for _, validator := range validators {
		return validator, nil
	}

	return nil, errors.New("unknown validator")
}
