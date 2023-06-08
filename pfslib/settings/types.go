package settings

type PfsSettings struct {
	EtcdSettings  etcdSettings  `yaml:"etcd,flow"`
	MinioSettings minioSettings `yaml:"minio,flow"`
}

type etcdSettings struct {
	Endpoints []endpoint `yaml:",flow"`
}

type minioSettings struct {
	Endpoint    endpoint         `yaml:"endpoint"`
	Credentials minioCredentials `yaml:"credentials,flow"`
}

type minioCredentials struct {
	AccessKeyId     string `yaml:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey"`
}

type endpoint struct {
	Ip   string `yaml:"ip"`
	Port int    `yaml:"port"`
}
