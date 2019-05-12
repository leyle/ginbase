package ginbase

type Config interface {
}


type IMongoDbConf interface {
	Url() string
	DbName() string
}

type IRedisConf interface {

}