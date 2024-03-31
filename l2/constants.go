package l2

const FileHeaderSize = 24
const CTMHeaderSize = 12
const DefaultMessageSize = 2432
const MessageHeaderSize = 16
const MessageBodySize = DefaultMessageSize - CTMHeaderSize - MessageHeaderSize
