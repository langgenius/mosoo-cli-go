# Public Thread Run Workflow

Use this workflow to create or continue a Public Thread, wait for final output,
or inspect a transcript.

```sh
mosoo public-thread-api threads create --agent-id <agent-id> --file body.json --wait -o json
mosoo public-thread-api threads create --agent-id <agent-id> --file body.json --final-output
mosoo public-thread-api events wait --thread-id <thread-id> --final-output
mosoo public-thread-api threads transcript --thread-id <thread-id>
```

Prefer `--wait` when the caller needs run status and structured output. Prefer
`--final-output` when only the final Agent answer should be returned. Use
`threads transcript` when the caller needs a readable event transcript for an
existing Thread.
