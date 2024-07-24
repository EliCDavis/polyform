# PLY Utils

Utils for interacting with PLY files

## Header

Writes out PLY header information

```bash
ply-utils -i .\out.ply header --json --out header.json
```

## Properties

### Remove Properties 

Below is an example on how you could remove the color components of a ply file

```bash
ply-utils -i in.ply property remove red green blue
```

### Add Properties

```bash
ply-utils -i in.ply property add red char green char blue char
```

### Analyze Properties

Builds summaries of data pertaining to the properties within the PLY file

```bash
# Analyze all properties in the file
ply-utils -i in.ply property analyze

# Analyze only red, green, and blue
ply-utils -i in.ply property analyze red green blue
```