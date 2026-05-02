Output docs/demo.gif

Set FontSize 24
Set Width 1200
Set Height 700
Set TypingSpeed 320ms
Set PlaybackSpeed 0.7
Set CursorBlink false

Env HOME "/tmp/aoo-vhs-home"
Env XDG_CONFIG_HOME "/tmp/aoo-vhs-home/.config"
Env AOO_NOTES_DIR "./examples/notes"
Env AOO_APP_DIR "/vhs"

Type "./bin/aoo"
Enter
Wait+Screen /notes>/
Sleep 5000ms

Type "ssh jump"
Wait+Screen /example ssh jump/
Sleep 6500ms

Right
Wait+Screen /RUN/
Sleep 5500ms

Enter
Wait+Screen /run\? \[y\/N\]:/
Sleep 8000ms

Type "n"
Enter
Wait+Screen /cancelled/
Sleep 5500ms

Type "./bin/aoo"
Enter
Wait+Screen /notes>/
Sleep 5000ms

Type "grep text"
Wait+Screen /grep text in files recursively/
Sleep 6500ms

Enter
Wait+Screen /\[template\] grep text in files recursively/
Sleep 5000ms

Type "ERROR|CRITICAL"
Enter
Sleep 4000ms
Type "/var/log"
Enter
Wait+Screen /run\? \[y\/N\]:/
Sleep 8000ms

Type "n"
Enter
Wait+Screen /cancelled/
Sleep 6000ms
