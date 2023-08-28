package tests

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"main/internal/cache"
	"main/pkg"
	"main/pkg/utils"
	"testing"
	"time"
)

func TestAddToCache(t *testing.T) {
	testCases := []struct {
		key        string
		value      string
		expiration int
	}{
		{
			key:        "hello",
			value:      "world",
			expiration: 5,
		},
		{
			key:        "hello1",
			value:      "",
			expiration: 5,
		},
		{
			key:        "1",
			value:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed nec justo eu mauris venenatis tincidunt. Nulla eget nunc et elit laoreet sollicitudin in vel risus. Fusce ultrices, justo et cursus suscipit, arcu nisi rhoncus augue, eu convallis sapien libero vel est. Vivamus vel metus nec purus venenatis bibendum nec ac leo. Duis vel neque nec tellus ultrices efficitur at nec eros. Etiam commodo nunc quis fermentum congue. Sed venenatis hendrerit metus eu dictum. Proin ut dui nec leo blandit gravida nec in elit. Cras pharetra justo sit amet nisl efficitur, ac rhoncus purus cursus. Vivamus nec bibendum justo. Suspendisse euismod lacus et nibh eleifend, at placerat tellus accumsan. In ultrices vehicula orci eu lacinia. Suspendisse potenti.\n\n\n\n\n",
			expiration: 5,
		},
		{
			key:        "",
			value:      "hello world",
			expiration: 5,
		},
		{
			key:        "",
			value:      "",
			expiration: 5,
		},
	}

	ctx := context.Background()
	cfg := utils.LoadConfig("../config/app.yaml")

	rdb, err := pkg.NewRedisClient(context.Background(), cfg)
	if err != nil {
		log.Fatalln("Error create redis client:", err)
	}

	for _, tc := range testCases {
		exp := time.Duration(tc.expiration) * time.Second
		require.NoError(t, cache.AddToCache(ctx, rdb, tc.key, tc.value, exp))

		data, err := rdb.Get(ctx, tc.key).Result()
		require.NoError(t, err)

		bytes, err := json.Marshal(tc.value)
		require.NoError(t, err)

		assert.Equal(t, bytes, []byte(data))
	}
}

func TestGetFromCache(t *testing.T) {
	testCases := []struct {
		key        string
		value      string
		expiration int
	}{
		{
			key:        "hello",
			value:      "world",
			expiration: 5,
		},
		{
			key:        "hello1",
			value:      "",
			expiration: 5,
		},
		{
			key:        "1",
			value:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed nec justo eu mauris venenatis tincidunt. Nulla eget nunc et elit laoreet sollicitudin in vel risus. Fusce ultrices, justo et cursus suscipit, arcu nisi rhoncus augue, eu convallis sapien libero vel est. Vivamus vel metus nec purus venenatis bibendum nec ac leo. Duis vel neque nec tellus ultrices efficitur at nec eros. Etiam commodo nunc quis fermentum congue. Sed venenatis hendrerit metus eu dictum. Proin ut dui nec leo blandit gravida nec in elit. Cras pharetra justo sit amet nisl efficitur, ac rhoncus purus cursus. Vivamus nec bibendum justo. Suspendisse euismod lacus et nibh eleifend, at placerat tellus accumsan. In ultrices vehicula orci eu lacinia. Suspendisse potenti.\n\n\n\n\n",
			expiration: 5,
		},
		{
			key:        "",
			value:      "hello world",
			expiration: 5,
		},
		{
			key:        "",
			value:      "",
			expiration: 5,
		},
	}

	ctx := context.Background()
	cfg := utils.LoadConfig("../config/app.yaml")

	rdb, err := pkg.NewRedisClient(context.Background(), cfg)
	if err != nil {
		log.Fatalln("Error create redis client:", err)
	}

	for _, tc := range testCases {
		exp := time.Duration(tc.expiration) * time.Second

		data, err := json.Marshal(tc.value)
		require.NoError(t, err)

		_, err = rdb.Set(ctx, tc.key, data, exp).Result()
		require.NoError(t, err)

		var s string
		require.NoError(t, cache.GetFromCache(ctx, rdb, tc.key, &s))

		assert.Equal(t, tc.value, s)
	}
}
