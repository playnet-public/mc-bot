package valheim

import (
	"context"

	a2s "github.com/rumblefrog/go-a2s"
)

// Info returns an a2s ServerInfo response
func (c Client) Info() (*a2s.ServerInfo, error) {
	return c.a2sClient.QueryInfo()
}

// CountPlayers on the Server right now
func (c Client) CountPlayers(ctx context.Context) (int, error) {
	playerInfo, err := c.a2sClient.QueryPlayer()
	if err != nil {
		return -1, err
	}

	return int(playerInfo.Count), nil
}

// Players on the server right now
// NOTE: Valheim currently does not provide the player names, so the list of players is left empty
func (c Client) Players(ctx context.Context) (int, []string, error) {
	playerInfo, err := c.a2sClient.QueryPlayer()
	if err != nil {
		return -1, nil, err
	}

	// playerNames := make([]string, 0, playerInfo.Count)

	// for _, player := range playerInfo.Players {
	// 	playerNames = append(playerNames, player.Name)
	// }

	return int(playerInfo.Count), []string{"<unknown>"}, nil
}
