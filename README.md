# Gravastar

Gravastar is a browser-based gravity game where players create “Dots” on a shared canvas visible to everyone around the world.
Each Dot has mass and interacts through gravity, so a Dot created in Germany can influence one in Australia.

Play online: https://gravastar.shivansh.io

## Running it yourself

Want to host your own instance? Here are two ways to do it.

### Option 1: Docker (recommended)

First, clone this repository and create your config file.

```bash
cp config/config.sample.json config/config.json
```
You can edit `config/config.json` if you want to change settings, but the defaults work fine.

Build and run:
```bash
docker build -f Containerfile -t gravastar:latest .
docker run -d -p 8080:8080 -v $(pwd)/config/config.json:/service/config/config.json gravastar:latest
```

Open http://localhost:8080 and you should see the game!

### Option 2: Go

If you have Go installed, you can run it directly:

Clone this repository and create your config file just as described in the Docker section above.

Build and run:
```bash
go build -o bin/gravastar cmd/gravastar/main.go
./bin/gravastar -config config/config.json
```

Then open http://localhost:8080

## Chaos Management

N-body simulations are chaotic, meaning tiny differences cause big changes over time.
So two browsers running the same simulation will drift apart pretty quickly.

To fix this, there's a master simulation running on the server.
Players can hit the "Sync" button to reset their local simulation to match the server's state.
