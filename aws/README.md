# cloud-uploader-lab


| Student | Mentor | Cloud |
| ------ | ------ | ------ |
| Vladimir Glushakov | Iurii Shakhmatov | GCP |
| Alexander Ryzhickov | Iurii Shakhmatov | GCP |
| Timur Akhmadiev | Iurii Shakhmatov | GCP |
| Vyacheslav Starostin | Iurii Shakhmatov | AWS |
| Maksim Meshkov | Nikita Ivanov | AWS |
| Denis Zakharov | Nikita Ivanov | Azure |
| Alexander Okhotin | Nikita Ivanov | Azure |

# How to run tests

```
make test
```

You need to have installed Docker and Make.

# Examples

See `_examples` folder.

# AWS configuring

AWS's `OpenBucket` uses default config, see [Specifying credentials](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials). For example you can set the env variables `AWS_REGION`, `AWS_ACCESS_KEY_ID`, and `AWS_SECRET_ACCESS_KEY`.
