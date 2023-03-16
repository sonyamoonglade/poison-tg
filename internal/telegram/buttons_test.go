package telegram

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectExtractMsgID(t *testing.T) {
	t.Parallel()
	type testcase struct {
		msgIDs   []int64
		callback int
	}
	var testcases []testcase
	for i := int64(-math.MaxInt16); i < math.MaxInt16; i++ {
		testcases = append(testcases, testcase{msgIDs: []int64{i, i + 1, i + 2}, callback: int(math.MaxInt16 - i)})
	}

	for _, tc := range testcases {
		data := injectMessageIDs(tc.callback, tc.msgIDs...)
		require.NotZero(t, data)
		gotMsgIDs, callback, err := parseCallbackData(data)
		require.NoError(t, err)
		require.EqualValues(t, tc.msgIDs, gotMsgIDs)
		require.Equal(t, tc.callback, callback)
	}
}
