# Public Thread File Upload Workflow

Use this workflow when a Public Thread needs file attachments.

Run the generated commands in order and save the returned `fileId` before moving
to the next step:

```sh
mosoo console-rest files create-upload --file upload.json -o json
mosoo console-rest files upload-content --file-id <file-id> --file content-body.json -o json
mosoo console-rest files complete-upload --file-id <file-id> --file complete.json -o json
mosoo public-thread-api files add --thread-id <thread-id> --set fileId=<file-id> -o json
```

Use `mosoo commands show <path...> --json` before each command to confirm body
shape and host selection. If a step fails or times out, inspect state with
`console-rest files get-upload` before retrying. Use
`console-rest files abort-upload` only for a pending upload that should not be
completed.
