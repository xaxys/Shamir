# Shamir

网安上机Shamir密钥分享算法, 获取到足够多秘钥后才能正确解出密文

- 先随机生成一个巨大的质数做模数

- 然后再生成一堆（t个）长长的多项式系数，代码里默认256位，常数项系数作为密文

- 根据多项式生成n个点，横坐标就取123...纵坐标算出来也是长长的是用于分享的秘钥

- 取t个秘钥，即t个点，就可以还原出拉格朗日多项式系数，此处直接把x=0带入求常数项系数就行了，也就是密文。

## Demo

生成密文

[![XUntQS.md.png](https://iili.io/XUntQS.md.png)](https://freeimage.host/i/XUntQS)

数量不足的秘钥

[![XUo93u.md.png](https://iili.io/XUo93u.md.png)](https://freeimage.host/i/XUo93u)

数量足够的秘钥

[![XUo3TQ.md.png](https://iili.io/XUo3TQ.md.png)](https://freeimage.host/i/XUo3TQ)

## Reference

特别感谢[FORIMOC](https://github.com/FORIMOC)的教程和拿来就用的JQuery前端
