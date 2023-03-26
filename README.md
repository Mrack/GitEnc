# GitEnc

GitEnc是一款可以对Git存储库的文件进行加密的软件，你可以提供自定义key或者随机生成key对文件进行加密。

## 使用方法

1. 初始化git仓库
``` 
cd repo
git init
```

2. 对git仓库进行配置
``` 
gitenc init -key <your password> 
```

3. 编辑.gitattributes匹配需要加密的文件
https://www.w3cschool.cn/doc_git/git-gitattributes.html

4. 提交代码到本地git缓存区
``` 
git add *
``` 

5. 推送代码到本地git库以及远程仓库
``` 
git commit -a -m “Init” & git push
```


