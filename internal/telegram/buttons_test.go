package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectExtractMsgID(t *testing.T) {
	type testcase struct {
		msgID    int64
		callback int
	}
	var testcases []testcase
	for i := 0; i < 1000; i++ {
		testcases = append(testcases, testcase{msgID: int64(i), callback: i})
	}

	for _, tc := range testcases {
		data := injectMessageID(tc.msgID, tc.callback)
		require.NotZero(t, data)

		msgID, callback, err := ExtractMsgID(data)
		require.NoError(t, err)
		require.Equal(t, tc.msgID, msgID)
		require.Equal(t, tc.callback, callback)
	}
}
