# wow-server-status

Sometimes it's patch day (every week in fact!)
and you want to play World of Warcraft,
but the servers are down.
This script uses the Blizzard API
to figure out when your server is back up.
Right now it only works for Area 52,
because that's where I play,
but it would be really easy to modify the code to work for your server :)

## Usage

1. **Get credentials**
  1. Go to https://develop.battle.net/ and accept their license.
  1. Create a new client (`+ Create Client`)
  1. Set `Client Name`, `Redirect URLs`, and `Intended Use` to whatever you want,
     so long as it complies with the Blizzard terms of use.
     We won't be using a redirect URL,
     so you don't have to worry about setting it to a specific value.
  1. Copy `Client ID` and put it in `secrets/client_id.txt`.
  1. Copy `Client Secret` and put it in `secrets/client_secret.txt`.
1. Run the program!
  1. ```shell
     go run main.go
     ```

## License

MIT Open Source. See [LICENSE](/LICENSE) for exact details.
