# fate (f8)

An file storage bucket system where abstract entities have file storage capabilties

Consists of an Entity manager where one can register an entity

```sh
git clone --recurse-submodules -j8 https://github.com/phanirithvij/fate
cd fate
# this will build filebrowser assets
bash custom-fb.sh -a -d
bash custom-fb.sh -r -d
go build
./fate # .\fate.exe
```

```sh
# ssh to heroku instance
heroku run bash
# ssh to container
heroku ps:exec
# list all builds
heroku builds
# cancel running build with id
heroku builds:cancel <id>
# cancel running build
heroku builds:cancel
# show logs
heroku logs --tail
```

## Usage (undecided)

Example

```go
// TODO
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
- Using [Polymorphic](https://gorm.io/docs/has_many.html#Polymorphism-Association) for the sql relations instead of foreignKey because there's some bugs with foreignKey schema generation in gorm (In my schema)
  - And in v2 of gorm the `db.Model(&model).AddForeignKey` method was removed

## TODO

- [x] Make gorm an internal dependency
  - [x] i.e. don't force developer to use gorm
  - [x] If they are using they can provide the gorm.DB instance
  - [x] If not we will point them to the default database
- [ ] Add Examples and tests
  - [ ] Make the main package a test file
  - [ ] Make the main package an example file
- [ ] UPX needs `--strip-relocs=0` for go executable uncompression to work
  - Example https://github.com/upx/upx/issues/2#issuecomment-730387507
  - `upx fate.exe --strip-relocs=0`
  - `upx -d fate.exe -o unc.exe --strip-relocs=0`
  - Even then diff will show different binaries
- [ ] Documentation
  - [ ] Mention the defualt values for all things
  - [ ] Mention it uses Sqlite or PostgreSQL and can be configured
- [ ] Remove entityMap and use a Cahce of some sort to make things efficient
  - Because it is ugly
  - [ ] Own cachemap for each entity instead of the global one?
- [ ] CI
  - [ ] Tests
  - No releases as just a library
- [ ] File Upload utitlites
  - [ ] Continuous upload (normal)
  - [ ] Tusd resumable upload along with uppy.js
- [ ] Web ui for managing entity files
  - [ ] Needed for my BTP project
    - [ ] find or create separate library ?
  - [ ] Mention PostgreSQL so db can also be managed
- [x] BoltDB ? Not suitable

# FAQ

## Why f8?

Abstract Entity File Transfer (aeft -> fate -> f8)

## How useful is this?

Not sure. Just made it for some of my projects which all require file uploads, exports, sharing and file transfer analytics
