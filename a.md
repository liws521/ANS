# FSE实现过程示例
- 本文档旨在用FSE对一个长度为8的字符串ACBCACBAB进行编码解码,借此具体阐述FSE的工作步骤细节,包括符号的标准化统计,编码表解码表的构建过程,以及在编码解码过程中的详细状态转换,借此加深对FSE工作细节的了解.

## 1.准备工作
### 1.标准化normalized
- 首先统计字符串中每个符号出现的频次,对字符串做normalizedCount处理,保证频次总数为2的n次幂,且出现次数>=1的符号的频次至少为1.
- 样例:
    - Symbol    A	B	C
    - Count	    3	3	2
- 根据normalizedCount统计的结果(以下统称N_table),确定状态表示位数tablelog=3,
- 表格大小 tableSize = 1 << tablelog = 8
### 2.符号区间与符号分布
- P(A)=3/8,-log_2 P(A)=1.41, 用1bit~2bits来编码
- P(B)=3/8,-log_2 P(B)=1.41, 用1bit~2bits来编码
- P(C)=2/8,-log_2 P(C)=2, 用2bits来编码
- 符号分布以3为步长,最后得到stateTable如下所示
    - stateTable	0	1	2	3	4	5	6	7
    - symbol	    A	B	C	A	B	C	A	B
- 接下来构建编码表
## 2.构建编码表
- 构建编码表的代码如下
```
for i in range(tableSize):
    s = stateTable[i]
    index = symbol_list.index(s)
    #symbol_list: ['A', 'B', 'C']
    #index是取每个字符在上表中的索引
    codingTable[cumulative[index]] = tableSize + i
    #cumulative [0, 3, 6, 8, 9],是根据N_table做的字符数量累积
    #根据符号在stateTable中出现的顺序,构建编码表
    cumulative[index] += 1
```
- 最后构建的编码表如下所示
    - codingTable	0	1	2	3	4	5	6	7
    - state	        8	11	14	9	12	15	10	13

编码表中的状态值将范围从[0 ~ tableSize-1]扩展到了[tableSize ~ 2*tableSize-1],后续算nbBitsOut等值时更加方便
最后还需要一个symbolTT符号转换表,用来保存每个符号的一些信息
构建规则

        maxBitsOut = tableLog - first1Index(occurrences - 1)
        minStatePlus = occurrences << maxBitsOut
        #出现次数左移最大位数是占用n位还是n+1位的分界线
        symbolTT[symbol]['deltaNbBits'] = (maxBitsOut << 16) - minStatePlus
        # 最大位数左移16位,减去分界线,这个值后来再加上状态值,再右移就能知道实际输出位数
        symbolTT[symbol]['deltaFindState'] = total - occurrences
        # 它出现之前的所有符号个数,减去它的个数
        total += occurrences

构建出来的表格

Symbol	A	B	C
deltaNbBits	131060	131060	196592
deltaFindState	-3	0	4

至此,我们已经根据N_table的信息构建出了编码所需的所有表单

3.构建解码表

利用N_table与stateTable构建解码表D_table

decodeTable = [{} for _ in range(tableSize)]
nextt = list(normalizedCount.items())
for i in range(tableSize):
    t = {}
    t['symbol'] = stateTable[i]
    index = symbol_list.index(t['symbol'])
    x = nextt[index][1]
    nextt[index] = (nextt[index][0], nextt[index][1] + 1)
    t['nbBits'] = tableLog - first1Index(x)
    t['base'] = (x << t['nbBits']) - tableSize
    decodeTable[i] = t

构建出来的解码表如下

D_table	0	1	2	3	4	5	6	7
symbol	A	B	C	A	B	C	A	B
nbBits	2	2	2	1	1	2	1	1
base	4	4	0	0	0	4	2	2

至此,编码表与解码表全都构建完毕,接下来举一个例子来进行编码到解码的全过程

4.编码

编码过程代码

def encodeSymbol(symbol, state, bitStream, symbolTT):
    nbBitsOut = (state + symbolTT['deltaNbBits']) >> 16
    bitStream += outputNbBits(state,nbBitsOut)
    state = codingTable[(state >> nbBitsOut) + symbolTT['deltaFindState']]
    return state, bitStream

编码表

codingTable	0	1	2	3	4	5	6	7
state	8	11	14	9	12	15	10	13
Symbol	A	B	C
deltaNbBits	131060	131060	196592
deltaFindState	-3	0	4

以字符串ACBCABAB为例,来模拟编码解码过程
首先运行一次encodeSymbol(A, 0, bitStream, symbolTT):
以0作为起始状态,进入编码状态
state = codingTable[(state >> nbBitsOut) + symbolTT['deltaFindState']]=15 

ibelee勘误: 这里有所出入,FSE源码是采用第一个符号的boundary作为初始状态进行压缩的

本例在调研阶段不了解的时候用了0作为初始状态




接下来开始进行编码
1.输入待编码字符与当前状态
2.通过 nbBitsOut = (state + symbolTT['deltaNbBits']) >> 16 计算出nbBitsOut
3.将State的低nbBitsOut位写入bitStream
4.通过state = codingTable[(state >> nbBitsOut) + symbolTT['deltaFindState']]计算出下一个状态
5.若仍有待编码字符,则读取下一个编码字符,回到步骤1;若所有字符编码完毕,将当前状态的低tablelog位写入bitStream

state	S	nbbits	bitstream	subr	index	state_n
15	A	2	11	3	0	8
8	C	2	1100	2	6	10
10	B	1	11000	5	5	15
15	C	2	1100011	3	7	13
13	A	2	110001101	3	0	8
8	B	1	1100011010	4	4	12
12	A	2	110001101000	3	0	8
8	B	1	1100011010000	4	4	12
12	end	tablelog	1100011010000100	
	
	


上表中为了美观,项名做了一定程度简写,在此进行说明
subr=State>>nbBitsOut
index=State>>nbBitsOut+deltaFindState
至此,字符串ACBCABAB已经完全编码完毕,编码后的bitStream为1100011010000100

5.解码
D_table	0	1	2	3	4	5	6	7
symbol	A	B	C	A	B	C	A	B
nbBits	2	2	2	1	1	2	1	1
base	4	4	0	0	0	4	2	2

接下来通过解码表对bitStream1100011010000100进行解码,代码如下

def decodeSymbol(state, bitStream, D_table):
    symbol = D_table[state]['symbol']
    nbBits = D_table[state]['nbBits']
    rest, bitStream = bitsToState(bitStream, nbBits)
    state = D_table[state]['base'] + rest
    return symbol, state, bitStream

第一步先从bitStream的末尾读取长度为tablelog的bit位,作为解码开始的第一个状态
开始按以上代码表述过程进行解码

state	bitstream	S	nbits	rest_b	rest_d	base	state_n
4	1100011010000	B	1	0	0	0	0
0	110001101000	A	2	00	0	4	4
4	1100011010	B	1	0	0	0	0
0	110001101	A	2	01	1	4	5
5	1100011	C	2	11	3	4	7
7	11000	B	1	0	0	2	2
2	1100	C	2	00	0	0	0
0	11	A	2	11	3	4	7
7	end	
	
	
	
	
	


由于先编码的后被解码,解码出来的符号,即为ACBCABAB,解码成功


