# B.O.B - The Boss Of Brevity

Meet B.O.B, the boss of brevity, the sultan of shortness, making your links famous for being short and sweet.

## What is B.O.B?
B.O.B is a URL shortener, which means it takes a long URL and makes it shorter. It's a simple concept, but it's a useful one. B.O.B is a free and open source URL shortener, which means you can use it for free and you can see how it works. B.O.B is also a self-hosted URL shortener, which means you can host it yourself and you don't have to rely on a third party to keep your links alive.

## Features
Some of the features B.O.B has include:

- Shorten URLs
- Advanced analytics
- API
- Editing short links
- Custom URLs
- Self-hosted
- Free and open source

## Future plans
Some of the things we plan to add to B.O.B in the future include:

- A web interface
- Expiration dates for short links
- A mobile app
- Several social media bots that will support multiple platforms
- Premium features
  - Password protected links

## Getting started
To get started with B.O.B, you'll need to install it.

- You can do this by downloading the latest release from the releases page (TODO).
- You can also build it yourself by cloning this repository and running `go
  build` in the root directory. Ensure you have Go installed if you want to do
  this. If Go is not installed, you can download it from
  [here](https://golang.org/dl/). Prefer version 1.20 or higher but it should
  work with older versions too.
- Or you can use our `Dockerfile`.

Ensure supported environmental variables have been set or provide them to the
`bob` executable via cli.

If not using `docker`, you can run it with `./bob` (or `bob.exe` on Windows).
This will start B.O.B on port 8080. You can change the port by setting the `PORT`
environment variable. For example, `PORT=3000 ./bob` will start B.O.B on port
3000.

## Configuration
B.O.B can be configured using environment variables. The following environment variables are available:

- `HOST`: The host to run B.O.B on. Defaults to `127.0.0.1`
- `PORT`: The port to run B.O.B on. Defaults to `8080`.
- `MONGODB_CONNECTION_URL`: The connection URL of the mongodb database to use. Required.
- `DEV_MODE`: Set to true if you want to run without a mongodb connection URL.
  B.O.B will use an in-memory db.

You can also use cli flags to provide configuration values. For example, `./bob
--dev` will start B.O.B in development mode.

If starting B.O.B using docker, set the `environments` values with your own
configuration or run it as it is.

**NOTE**: `MemDB` is not as restrictive and robust as the `MongoDB`
implementation. If you encounter any issues using it in dev mode, please create
a new issue.

## API
B.O.B has an API which can be used to interact with it. The API is documented in our [OpenAPI spec](./api.yaml).

## Contributing
Contributions are welcome! Please read our [contributing guidelines](./CONTRIBUTING.md) for more information.

## API Testing
To test the API, you can upload the [OpenAPI Specs](./api.yaml) we have provided to any API Testing Platform like Postman.

Or without leaving VSCode, you can use the [Thunder
Client](https://marketplace.visualstudio.com/items?itemName=rangav.vscode-thunder-client)
extension to test the API. Click on the `Collections` tab on the left sidebar
and click on the `Import` button. Then select the `api.yaml` file. That's it!
You can now test the API.

## Donations
If you like B.O.B and want to support its development, you can donate to us using the following methods:

- [Buy me a coffee](https://bmc.link/philemon)

## License
B.O.B is licensed under the [MIT License](./LICENSE).

## Credits
B.O.B is developed and maintained by [Philemon Ukane](github.com/ukane-philemon). It is inspired by [bit.ly](https://bit.ly) and [tinyurl.com](https://tinyurl.com). It is built using [Go](https://golang.org) and [MongoDB](https://mongodb.com). It is hosted on [Render](https://render.com) (for free!).

## Contact
You can contact me via email at [ukanephilemon@gmail.com](mailto:ukanephilemon@gmail.com) or on [Twitter](https://twitter.com/behindtextdev).

PS: I not always available on Twitter, so email is the best way to reach me.
PSS: I was unable to put in much time into this project, so it's not as good as I would have liked it to be. I hope to improve it in the future.

