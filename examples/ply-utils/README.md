# PLY Utils

Utils for interacting with PLY files

## Header

```
ply-utils -i .\out.ply header --json --out header.json
```

## Properties

### Remove Properties 

Below is an example on how you could remove the color components of a ply file

```
ply-utils -i in.ply property remove red green blue
```

### Add Properties

```
ply-utils -i in.ply property add red char green char blue char
```