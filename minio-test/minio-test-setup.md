# MinIO server setup

1. Downloading MinIO server
```
wget https://dl.min.io/server/minio/release/linux-amd64/minio

chmod +x minio

MINIO_ROOT_USER=admin MINIO_ROOT_PASSWORD=password ./minio server /mnt/data --console-address ":9001"
```

> Note: replace `/mnt/data` with the desired data location (better if persistent location)

<br>

2. Go to MinIO user interface through `localhost:9000`, log in with user and password previously set and dive into `Access Keys` submenu.  
Create an Access Key and copy `Access Key` and `Secret Key` values, you will need them for MinIO Client setup.

<br>

# MinIO client setup

1. Open `main.go` file and search for `getMinioClient` function; once there, replace `accessKeyID` and `secretAccessKey` values with your own previously created through MinIO user interface.

2. Simply run the program using `go run main.go` to create a bucket called 'testbucket' and store a text file in the bucket.

3. Check whether the uploading process was successful by navigating to the bucket using MinIO user interface, through `Object Browser` menu.