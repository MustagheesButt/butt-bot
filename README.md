Rename `conf.example.yml` to `conf.yml` and fill it in.

See [Opus dependencies](https://github.com/hraban/opus?tab=readme-ov-file#build--installation)

opusfile.pc may not be found on macOS, so instead of `brew install opusfile` try `sudo port install opusfile` using MacPorts.

Need `ffmpeg` and `https://github.com/bwmarrin/dca/tree/master/cmd/dca`. Make sure `~/go/bin` is in your path.

Also need `neofetch` and `sed` for `info` command.

## Todos

- Try youtube-dlp instead