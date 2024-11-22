# SPZ File Format

Niantic Scaniverse's [SPZ format](https://scaniverse.com/news/spz-gaussian-splat-open-source-file-format)

## API

### Read

Deserialize a gaussian splat from the input reader.

```go
spz.Read(in io.Reader) (*spz.Cloud, error)
```

### Load

Opens the file located at the `filePath` and deserializes a gaussian splat.

```go
spz.Load(filePath string) (*spz.Cloud, error)
```

### Read Header

Deserialize a gaussian splat from the input reader.

```go
spz.ReadHeader(in io.Reader) (*spz.Header, error)
```