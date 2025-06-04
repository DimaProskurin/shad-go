//go:build !solution

package retryupdate

import (
	"errors"
	"github.com/gofrs/uuid"
	"gitlab.com/slon/shad-go/retryupdate/kvapi"
)

func UpdateValue(c kvapi.Client, key string, updateFn func(oldValue *string) (newValue string, err error)) error {
	var authErr *kvapi.AuthError
	var conflictErr *kvapi.ConflictError

LoopGet:
	for {
		getReq := kvapi.GetRequest{Key: key}
		getResp, err := c.Get(&getReq)
		switch {
		case errors.As(err, &authErr):
			return err
		case errors.Is(err, kvapi.ErrKeyNotFound):
			break
		case err != nil:
			continue LoopGet
		}

	LoopSetWithCalc:
		for {
			var oldValue *string
			var oldVersion uuid.UUID
			if getResp != nil {
				oldValue = &getResp.Value
				oldVersion = getResp.Version
			}
			newValue, err := updateFn(oldValue)
			newVersion := uuid.Must(uuid.NewV4())
			if err != nil {
				return err
			}

			setReq := kvapi.SetRequest{
				Key:        key,
				Value:      newValue,
				OldVersion: oldVersion,
				NewVersion: newVersion,
			}

		LoopSet:
			for {
				_, err = c.Set(&setReq)
				switch {
				case errors.As(err, &authErr):
					return err
				case errors.As(err, &conflictErr):
					if conflictErr.ExpectedVersion == newVersion {
						return nil
					}
					continue LoopGet
				case errors.Is(err, kvapi.ErrKeyNotFound):
					getResp = nil
					continue LoopSetWithCalc
				case err != nil:
					continue LoopSet
				}
				return nil
			}
		}
	}
}
