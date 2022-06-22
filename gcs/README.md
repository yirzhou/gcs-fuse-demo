# Package GCS

## ObjectAttrs

```go
// Example ObjectAttrs
&storage.ObjectAttrs
{
    Bucket:"yirzhou",
    Name:"sample_folder/randfile",
    ContentType:"application/octet-stream",
    ContentLanguage:"",
    CacheControl:"",
    EventBasedHold:false,
    TemporaryHold:false,
    RetentionExpirationTime:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
    ACL:[]storage.ACLRule(nil),
    PredefinedACL:"",
    Owner:"",
    Size:104857600,
    ContentEncoding:"",
    ContentDisposition:"",
    MD5:[]uint8{0x2f, 0x28, 0x2b, 0x84, 0xe7, 0xe6, 0x8, 0xd5, 0x85, 0x24, 0x49, 0xed, 0x94, 0xb, 0xfc, 0x51},
    CRC32C:0xe6f38b62,
    MediaLink:"https://storage.googleapis.com/download/storage/v1/b/yirzhou/o/sample_folder%2Frandfile?generation=1655844885596653&alt=media",
    Metadata:map[string]string(nil),
    Generation:1655844885596653,
    Metageneration:1,
    StorageClass:"STANDARD",
    Created:time.Date(2022, time.June, 21, 20, 54, 45, 624000000, time.UTC),
    Deleted:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
    Updated:time.Date(2022, time.June, 21, 20, 54, 45, 624000000, time.UTC),
    CustomerKeySHA256:"",
    KMSKeyName:"",
    Prefix:"",
    Etag:"CO2T1vG2v/gCEAE=",
    CustomTime:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
}
```
