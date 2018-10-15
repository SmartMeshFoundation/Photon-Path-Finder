package SmartRaiden_Path_Finder
/*
==========================================
1、余额证明
PUT 	/<peerAddress>/balance

参数：1、
          	{
“balance_proof”:
{
“nonce”:					1234,
“transferred_amount”:		88,
“channel_id”:			5555,
“locksroot”:					“<..hash>”,
“additional_hash“:			“<...hash>”,
“signature”:				“对方的<signature>”
 },
 “locks”:[
  {
     “locked_amount”:			10,
 “expiration”:				200,
  “secret_hash”:			“<...hash>”,
                  },
{
     “locked_amount”:			10,
 “expiration”:				200,
  “secret_hash”:			“<...hash>”,
                  },]
  ...（所有的锁来验证locksroot）
加上alice的签名
如果有5000万的transferred_amount？如何证明
}
返回1：{“result”:”ok”}
返回2：{“result”:”invalid balance proof”}
×××××××××××××××××××××××××××××××××××××××
伪码：
onHandleBalanceProofMessage:
1\验证channel_id
check_channel_id(balance_proof.channel_id)
2\更新balance
Update_balance(channel_id,signature,nonce,
transferred_amount,sum_locked_amount(所有被锁的钱的总和))
{
//有一个ChannelToAddr,map[channel_id][]int{该通道参与者1Addr，该通道参与者2Addr}
//participant1,participant2=ChannelToAddr(key=channel_id)
//singer=ValidateSignature(signature)//消息是谁的签名
//发送此消息的人receiver=singer对比participant1,participant2},都不是,退出
//channel_alice=Graph[singer][receiver][“content”](1是通道的p1)
“content”包括：（只是数据结构）
selfAddr,
partnerAddr,
deposit（我存到通道的钱）
transferred_amount（我转给对方的钱）
,received_amount（对方转给我的钱）,
locked_amount(我被锁定的钱)，
费率（0.1%），
capacity(init=deposit我在此通道的可用余额)，
state(此通道他的状态open,close）,
balance_proof_nonce(nonce交易序号)
//channel_bob=Graph[receiver][singer][“content”]
//if nonce<=channel_alice.balance_proof_nonce retrun 过期的balance_proof
//channel_alice.更新可用余额(nonce,transferred_amount,sum_locked_amount)
//channel_bob.更新可用(nonce,transferred_amount,sum_locked_amount)
}
更新节点可用余额(nonce,transferred_amount,sum_locked_amount)
{
//判断nonce<=balance_proof，return //nonce比上一次记录的余额证明nonce还小
//capacity（我的可用余额）=deposit+received_amount-(transferred_amount+被锁定的钱)
}
//证明transferred_amount的合法 （计算两方deposit之和<双方的deposit（还得考虑锁定的钱））

==========================================
*/