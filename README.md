# gordle

A small terminal-based Wordle client written in Go.

## Screenshots

![Gameplay](screenshots/gameplay.png)

![Won](screenshots/won.png)

![Lost](screenshots/lost.png)

## Installation

If you have Go installed:

```bash
go install github.com/bur4ky/gordle@latest
```

You can also download a prebuilt version from the [latest release](https://github.com/bur4ky/gordle/releases/latest).

## Controls

Enter letters using your keyboard and press `Enter` to submit a guess. Use `Backspace` to remove a letter. Press `Ctrl+N` to start a new game, or `Esc` / `Ctrl+C` to exit.

Your daily game is saved automatically, so you can close gordle and return to it later without losing your progress. Statistics and streaks are saved as well.

When a new New York Times Wordle of the Day is available, gordle uses it. If it can't be reached, gordle falls back to a local list of answers.

## License

See [LICENSE](LICENSE).
