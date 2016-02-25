package goredis

const (
	REDIS_REPLY_STRING  = 1
	REDIS_REPLY_ARRAY   = 2
	REDIS_REPLY_INTEGER = 3
	REDIS_REPLY_NIL     = 4
	REDIS_REPLY_STATUS  = 5
	REDIS_REPLY_ERROR   = 6
)

type RedisReply struct {
	Type     int           /* REDIS_REPLY_* */
	Integer  int64         /* The integer when type is REDIS_REPLY_INTEGER */
	Len      int           /* Length of string */
	Str      string        /* Used for both REDIS_REPLY_ERROR and REDIS_REPLY_STRING */
	Elements int           /* number of elements, for REDIS_REPLY_ARRAY */
	Element  []*RedisReply /* elements vector for REDIS_REPLY_ARRAY */
}

func NewRedisReply(re interface{}, err error) *RedisReply {
	reply := &RedisReply{}
	if err != nil {
		reply.Type = REDIS_REPLY_ERROR
		reply.Str = err.Error()
		reply.Len = len(reply.Str)
		return reply
	}
	if re == nil {
		reply.Type = REDIS_REPLY_NIL
		return reply
	}
	switch re.(type) {
	case []uint8:
		reply.Type = REDIS_REPLY_STRING
		reply.Str = string(re.([]uint8))
		reply.Len = len(reply.Str)
	case []interface{}:
		reply.Type = REDIS_REPLY_ARRAY
		reply.Elements = len(re.([]interface{}))
		replys := make([]*RedisReply, reply.Elements)
		for i, r := range re.([]interface{}) {
			replys[i] = NewRedisReply(r, nil)
		}
		reply.Element = replys
	case int64:
		reply.Type = REDIS_REPLY_INTEGER
		reply.Integer = re.(int64)
	}
	return reply
}
