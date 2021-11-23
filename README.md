# HTTP-FETCH
This is a small tool designed to scrape one or more URLs given as command arguments.

## Usage
```http-fetch [--metadata] ...URLs```

The output files will be found under a folder called "output" under the root of this project's directory structure.

## Example
```http-fetch --metadata https://www.google.com```

## Docker Usage
```
docker build -t http-fetch .
docker run -v /Users/daniel/http-fetch-output:/app/output http-fetch --metadata https://www.google.com```
```

In the above run case, you may replace the ```/Users/daniel/http-fetch-output``` portion with any path on your local filesystem.

# TODO
- Add tests based around mocked web server. Test would ensure that file contents match between on-disk sample and scraped results
- Add support for nested resources loaded via JS and CSS
