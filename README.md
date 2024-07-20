Rename `conf.example.yml` to `conf.yml` and fill it in.

## Requirements

1. Install [Opus dependencies](https://github.com/hraban/opus?tab=readme-ov-file#build--installation)

    If `opusfile.pc` is not found on macOS, try `sudo port install opusfile` instead of `brew install opusfile`, using MacPorts.

2. Rename one of the binaries, according to your os/arch, in `/bin` to `dca` and put it in your PATH.

    **OR** Install [bwmarrin/dca](https://github.com/bwmarrin/dca/tree/master/cmd/dca). Make sure `~/go/bin` is in your PATH.

3. Install `ffmpeg` and `yt-dlp`

Also need `neofetch` and `sed` for `info` command.

## Todos

- ~~stop music~~
- ~~queue music, skip entry in queue~~
- make sure everything is non blocking
- paging in `list` command (or remove `list` entirely)
- ~~implement slash commands~~
- integrate pocketbase for song stats
