Practical lossless compression with latent variables using bits back coding
===

## meta
- 有一个Sender和Receiver想要进行通信, 通信的内容是一串字符序列seq, 为降低通信代价, 需要对seq进行数据压缩.
- 传统的方式是通过对seq的每个字符进行count, 得到一个定义在字符集A上的概率模型P, 基于这个概率模型与香农的信源编码定理对seq进行区间编码or熵编码, 当然也要将概率模型信息加在流的前面. Receiver收到流后先读取概率模型P, 而后依此进行解码, 得到原文.
- 如果不想对seq进行统计, 如何对其进行压缩呢?
- 对应seq中的一个字符s0, 在传统的方式下我们需要先知道它的频率P(s0), 现在我们不想做这个统计工作, 而是有S/R双方共享一份训练集, 对训练集进行抽样得到一个隐变量y0, 双方都能知道P(y0)和P(s0|y0), 用P(y0)编码隐变量y0, 用P(s0|y0)编码字符s0, 此时编码一个字符s0的长度为$-(logp(y0) + logp(s0|y0))$bits
- 到目前为止我们做到了什么呢, 做到了不对待压缩的seq进行概率统计就能压缩s0, 但此时的缺点是压缩一个字符所需的代价太大了
- 既然y0的获取方式并不重要, 那么我们可以从之前的压缩信息中得到y0, 以减少编码代价, 假设在压缩第一个字符s0之前给定一些额外位, 将这个位的一部分通过概率P(y)的解码器得到一个符号y0, 用这个y0下的P(s0|y0)来编码s0, 最后再用P(y0)把这个y0编码进去, 压缩过程进行了一次解码与两次编码, 先消耗了一些位, 而后又增加了一些位,这也是为什么这种思想被叫做`bits-back`, 总的编码s0的代价为
$$ -log p(s0|y0) - logp(y0) + logp(y0) = - logp(s0|y0)$$
- 到了这里我们把编码s0的代价优化了一次, 但仍然没有达到编码s0的理论熵值, 根据信源编码理论可知编码s0的理论熵值是$-log p(s0)$, 由贝叶斯公式反推出只有对y0的解码器的概率模型为P(y0|s0)时才能得到这个最佳编码理论值
- 这个概率我们没法用传统手段解决, 但是可以通过机器学习的方式训练出一个满足这个概率模型的生成模型, 训练出一个这样的编/解码器q(y|s), 以达到我们的目的
- 编码过程(假设s3已经编码完毕, 现在得到一个流str3)
    - 将str3的一些bits传入解码器生成模型q(y|s), 得到一个隐变量y4
    - 用概率模型P(s|y)的编码器编码符号s4
    - 用概率模型P(y)的编码器编码隐变量y4, 最终得到str4
    - 一次解码(消耗bits), 两次编码(增加bits), log下的加减为P下的乘除, 由贝叶斯公式得编码代价为$-log P(s4)$
- 解码过程(假设s5已经解码完毕, 现在得到一个流str4)
    - 用概率模型P(y)解码器解码隐变量y4
    - 用概率模型P(s|y)解码器解码符号s4
    - 将y4传入编码器生成模型q(y|s), 返回的bits添加到流上得到最终的str3
    - 与编码的完全逆过程, 两次解码一次编码
- 编/解码器如何选择的问题, 从上述过程可以看出在编码一个符号的过程中既有小的编码又有小的解码, 所以整体工作流程是FILO的stack-like模型, 在ANS理论出现之前, 通常把AC编码器的queue-like方式外包一层处理改为stack-like模式以配合BB的使用, 这当然带来了大量的额外性能损失, 而ANS理论的压缩流都是天然stack-like的, 与BB无缝衔接, 天作之合, 由此BB-ANS问世.



## Abstract
深度潜变量模型最近在许多数据领域取得成功.
无损压缩是这些模型的一项应用, 尽管具有很高的潜力, 但目前还没有实际实现.
我们提出BB-ANS, 一种用潜变量模型以几乎最优压缩率实现无损压缩的解决方案.
我们通过使用VAE压缩MINST数据集证明这个解决方案.

## 1 Introduction
信息理论和机器学习一直被视作遥不可及, 但其实两个领域很近, 是一个硬币的两面.
一项尤其优雅的连接是数据概率模型和无损压缩算法之间的连接.
香农的信源编码定理被认为是这种观点的基础, 并且在给定概率模型时有很多无损压缩算法的实现, huffman, AC, ANS.

隐变量与无损压缩之间 Bits back coding

我们的方案就压缩率和代码复杂度在现有的BB编码实现上加以改善, 允许用深度潜变量模型对任意大数据集进行高效无损压缩.
我们用VAE压缩MNIST数据集展示BB-ANS的效率, VAE是一种连续深度潜变量模型.
就我们所知, 这是BB第一次用连续潜变量模型实现.

## 2 Bits back coding
在这部分, 描述一下BB编码, 一种使用潜变量模型进行无损压缩的方法.
在描述BB本身之前, 先简短的讨论一下在给出一个完整的观测模型的基础上的编码, 这通常指区间编码/熵编码.
我们不描述这些算法的细节和实现, 而是描述高层特性以便更好理解BB算法.

为了简洁性, 下面部分的log2用log代替, 长度测量以bits为单位.

### 2.1 Compressing streams with AC vs. ANS
一个sender, 一个receiver, 想发送一个sequence, 以尽可能少的位数, 这个sequence中的符号都在一个字符集A中.
假设双方之前已知概率模型.

AC和ANS就是解决上述问题的, 提供从sequence(s)到比特流(message)的编码, 和比特流到还原数据流的解码.
AC和ANS的message length都是信息熵加上一个小的常量.
根据香农的信息熵理论, 长度不会低于这个值, 只会接近这个值.
对于长的数据流, 这个小常量被摊还(amortize)并且对压缩率的影响可以忽略不计.

AC和ANS的不同之处, AC是FIFO, queue-like, ANS是LIFO, stack-like.

只要给定一个概率分布p, AC/ANS的encoder/decoder的前后映射过程都是可逆的.

### 2.2 Bits back coding
现在给出BB编码的一个简短描述.
有关BB的更多起源过程看附录A.
我们假设可以访问一种编码方案, 可以根据任意分布编/解码的.
fjlakksdjf, 这里有句话, 重看

假设现在一个sender想要向receiver发送一个符号s0, 并且双方都可以用一个潜变量y访问生成的概率模型.
现在暂时我们让y是离散的, 2.5.1中我们再讨论两虚的潜变量.
假设双方都可以计算正向概率p(y) and p(s|y), 并且可以访问一个近似的q(y|s).
BB编码运算双方高效的对符号s0进行编解码.

必须假设伴随着样本s0, sender还有一些额外的bits要发送.
sender可以对这些extra bits进行解码来生成一个样本y0 ~ q(y|s0).
然后他们可以根据p(s|y0)编码s0, 和根据p(y)编码潜变量.
receiver做逆操作来还原潜变量和符号.
extra bits也可以被receiver还原, 通过根据q(y|s0)编码潜变量.
我们可以写出message length的增长, 一个公式

这个数量等于the negative of evidence lower bound(ELBO), 有时被称为模型的free energy.
[证据下界](https://blog.csdn.net/qy20115549/article/details/93074519)

### 2.3 Changing bits back coding
如果我们想编码一个数据点序列, 我们可以随机采样获得第一个数据点的extra bits.
然后使用编码后的第一个数据点作为第二个个数据点的extra information, 编码后的第二个数据点作为第三个数据点的extra information, and so on.
这种daisy-chain-like的方案首先由Frey(1997)所描述, 被称为'bits-back with feedback'.
我们简单的称之为chaining.

As Frey(1997) notes, 正如他所指出的, chaining不能直接由AC实现, 因为其数据的解码顺序.
Frey使用了一种stack-like的AC包装版本, 这招致了额外的代码复杂度和影响了压缩率.
压缩率的开销是因为AC在BB的每次迭代必须被flushed, 这又2-32的额外开销.

### 2.4 Chaining bits back coding with ANS
这部分主要讲了什么呢, 前面说了chaining与AC的问题是由于解码顺序, chaining可以直接和ANS配合而每次迭代没有额外开销.
这是因为ANS是天然的stack-like, 解决了AC和BB直接存在的问题.
现在描述这种新颖的算法, 我们称之为BB-ANS.

这部分最好对照着论文来看, 有很多图.
用一条线来象征栈顶或者说编解码的结尾.
当我们编码一个符号s到栈上的时候直接加在后面,栈变得更长了.
当我们从栈解码一个符号, 栈变得更短了, 并且得到了一个符号.

Table 1展示了sender的栈, 当使用BB-ANS算法编码一个样本, 从一些extra information和符号s0开始编码.
这个图太好了.好好看看.

这个过程明显是可逆的, 通过逆转操作的顺序, 并且用解码代替编码, 用编码代替采样.
此为这个过程是可以重复的, ANS栈在编码后仍然是一个ANS栈, 因此它可以被当作extra information来编码下一个符号.
这个算法可以兼容任何模型, whose prior, likelihood and (approximate) posterior 能被ANS编码的.
BB-ANS的一个简单的Python实现在附录C中被提供.

### 2.5 Issues affecting the efficiency of BB-ANS
很多因素能影响BB-ANS的效率, 

#### 2.5.2 The need for 'clear' bits
## 3 Experiments

### 3.1 Using a VAE as the latent variable model
我们使用VAE来论证BB-ANS编码方案.
这个模型有一个带有标准高斯分布先验概率和对角线高斯分布近似后验概率的多维隐变量.
我们选择一个输出分布, 似然p(s|y), 符合我们建模的数据领域.
通常VAE的训练目标是ELBO, 我们在2.2提过.
因此我们可以训练一个VAE并把它插进BB-ANS框架.

### 3.2 Compressing MNIST
我们考虑压缩MNIST数据集的任务.
首先用训练集训练一个VAE, 然后用训练好的VAE和BB-ANS压缩测试集.

MNIST数据集包含范围在整数0, ..., 255的整数.
除了压缩原始的数据集外, 我们也给出压缩二值化后的数据集的结果.
两个任务都使用VAE和一堆看不懂的东西.

## 5 Conclusion
数据的概率模型化是机器学习中的高度活跃领域.


## Appendix
### A: Bits back coding
我们在这给出BB算法的详细起源.

一如既往, 假设双方想发送一个符号s0, 并且双方都能访问隐变量生成模型y.
假设双方都能计算概率p(y)和p(s|y), 他们可能如何传递s0?

很简单, sender从p(y)中抽样得到y0, 然后用p(y)编码y0, 用p(s|y0)编码s0.
接将会导致message length为$-(logp(y0) + logp(s0|y0))$bits.
receiver将会根据p(y)首先从解码出y0, 然后根据p(s|y0)解码出s0.
然而还能最短更好, 能够有效的减少编码信息的长度.

首先如果sender还想发送给receiver一些额外的信息, 我们将使这个成为我们的优势.
我们假设其他信息使任意位的形式.
只要有足够多的位, sender能使用他们从p(y)中生成一个样本y0, 通过decode一些位.
- 这部分是怎么做? 真的用ANSdecode么

生成这个样本使用-logp(y0)位.(消耗掉了位数).
sender可以使用整形模型编码y0和s0, message length还是会增加那么多.
但是现在receiver可以首先解码s0和y0, 然后编码y0,恢复那些额外信息.
反转解码的过程以产生y0, 这也是bits back的由来.
这意味着编码s0的销毁变成了
$$ -log p(s0|y0) - logp(y0) + logp(y0) = - logp(s0|y0)$$

第二, 注意sender可以从任意分布中产生y0, 不必非要从p(y), 这可以是一个s0的函数.
如果我们生成模型并且用q(.|s)代表我们选择的分布, 可能会依赖s0.
数量等于ELBO.

这是信源编码理论的最佳长度.
因此BB能达到最佳压缩率, 如果sender和receiver能够访问这个后验概率.
通常找不到这样的后概率, 但是我们可以解决.

我们注意到两个人通过从后验概率中生成隐变量, 


### B: Discretization
正如我们在2.1讨论的一样, ANS是定义在有限符号集上的编码方案.
我该我们希望编码连续的变量, 就需要把它限定在有限符号集上.
这相当于离散化连续隐变量空间.

选择我们的离散化时, 重要的是下面几点.

### C:BB-ANS Python implementation
```python
def append(message, s):
    # (1) Sample y according to q(y|s)
    #       Decreases message length by -log q(y|s)
    message, y = posterior_pop(s)(message)

    # (2) Encode a according to the likelihood p(s|y)
    #       Increases message length by -log p(s|y)
    message = likelihood_append(y)(message, s)

    # (3) Encode y accoeding to the prior p(y)
    #       Increases message length by -log p(y)
    message = prior_append(message, y)
    return message

def pop(message):
    # (3 inverse) Decode y according to p(y)
    message, y = prior_pop(message)

    # (2 inverse) Decode s according to p(s|y)
    message, s = likelihood_pop(y)(message)

    # (1 inverse) Encode y accodeing to q(q|s)
    message = posterior_append(s)(message, y)
    return message, s
```

pop方法的每行都与append方法互逆.
