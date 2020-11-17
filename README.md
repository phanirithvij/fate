# fate (f8)

An file storage bucket system where abstract entities have file storage capabilties

Consists of an Entity manager where one can register an entity


## Usage (undecided)

```bash
go get -u -v "github.com/phanirithvij/fate/..."
```

Example

```go
t
f8
```

## Use cases

- Whenever an entity requires some file storage along with it this library can be used

**Some examples**

- A central Organization site where each organization stores files based on some criteria say for eg. user uploaded files. `Entity is Organization`
- A user collections system where user can upload files and need to have granular file permission control, Example some wallpaper website where each User has collections, uploaded images
- Can be used as a generic cloud file storage where there's only a single `Admin entity`
    - Extereme use cases are Imgur, Deviantart, Google Cloud, DropBox


**Notes**
- Will try to implement with as much abstraction as possible
- Most likely works alongside with `gorm`

# FAQ

## Why f8?

Abstract Entity File Transfer (aeft -> fate -> f8)

## How useful is this?

Not sure. Just made it for some of my projects which all require file uploads, exports, sharing and file transfer analytics
