package notify

// paceMessages holds 10 tiers × 10 copy variants for the hourly pace alert.
// Index 0 = tier 1 (projected remaining 0-10%, excellent).
// Index 9 = tier 10 (projected remaining 90-100%, catastrophic).
var paceMessages = [10][10]string{
	// Tier 1: 0–10% projected remaining — calm, no emoji
	{
		"完美節奏，繼續。",
		"用量效率極佳，維持現狀。",
		"本週使用率良好，無需調整。",
		"進度符合預期，保持就好。",
		"步調穩健，繼續。",
		"表現優秀，不需要改變任何事。",
		"使用率正常，照這樣走。",
		"額度使用充分，繼續保持。",
		"本週進度良好。",
		"用量達標，繼續。",
	},
	// Tier 2: 10–20% projected remaining — light encouragement, no emoji
	{
		"快用完了，最後衝一下。",
		"進度稍有剩餘，再來幾個任務收尾。",
		"接近滿載，稍微再多用一點。",
		"差一點達標，不要停。",
		"本週表現不錯，收個尾。",
		"最後一段路，把剩下的用掉。",
		"很接近了，給它一個完美收場。",
		"額度快見底，繼續就好。",
		"再多問 Claude 幾個問題就差不多了。",
		"快到了，再推一把。",
	},
	// Tier 3: 20–30% projected remaining — gentle push, no emoji
	{
		"稍微慢了一點，今天多交幾個任務給 Claude。",
		"速率偏低，把手邊能交給 Claude 的事丟給它。",
		"有點剩，今天多讓 Claude 幫你幹活。",
		"不急但要注意，稍微加快一點。",
		"別讓 Claude 沒事做，今天多問幾題。",
		"有些額度在閒著，想辦法用掉它。",
		"隨手問問 Claude 也好，別浪費。",
		"今天多開幾個 session，補一下進度。",
		"速率偏慢，多做幾件能用到 Claude 的事。",
		"有點懶，今天給自己一點目標。",
	},
	// Tier 4: 30–40% projected remaining — firm, no emoji
	{
		"落後了，快打開 Claude 開始幹活。",
		"進度不行，今天要把速率補上來。",
		"你最近在摸魚嗎？要開始動了。",
		"有不少額度在閒置，這不應該發生。",
		"你訂了 Claude Pro 卻沒好好用，有點可惜。",
		"今天要認真一點，把能交給 Claude 的全部丟給它。",
		"落後的不少，現在就去開一個新任務。",
		"你有 Claude Pro 卻讓它閒著，很浪費。",
		"認真補一下，今天的目標是多用 10%。",
		"進度需要修正，現在開始。",
	},
	// Tier 5: 40–50% projected remaining — emotional, no emoji
	{
		"你的錢快白花了，認真一點。",
		"快一半的額度要浪費掉了，你在幹嘛？",
		"照這個速率，你的訂閱費有一半在打水漂。",
		"醒一醒，你的 Claude Pro 額度在凋零。",
		"用量嚴重偏低，今天要補回來。",
		"你付了錢卻在浪費，有點可惜。",
		"Claude 在等你，你在哪裡？",
		"你的訂閱費正在蒸發，快行動。",
		"有夠慢，照這樣下去虧大了。",
		"快去把手邊的事交給 Claude，不然這週就廢了。",
	},
	// Tier 6: 50–60% projected remaining — frustrated + emoji
	{
		"😤 超過一半的額度要浪費了，你到底在搞什麼？",
		"💸 你的訂閱費照這個速率要沖馬桶了。",
		"😒 你有在用 Claude 嗎？我懷疑。",
		"🙄 What are you doing？額度在消失你還不動？",
		"😩 幹，你付了多少錢，現在一半以上在浪費。",
		"💀 你是真的不知道還是假裝不知道自己在浪費錢？",
		"😤 照這個速率，你這週基本上在捐款給 Anthropic。",
		"🚨 嚴重落後，你現在應該立刻打開一個新任務。",
		"😡 用量低到我以為你取消訂閱了。",
		"💸 你知道你在幹嘛嗎？你在把錢丟進海裡。",
	},
	// Tier 7: 60–70% projected remaining — angry + profanity
	{
		"🤬 你到底在搞什麼鬼，這個速率完全不行。",
		"😤 Seriously, what the f**k are you doing？額度都在跑。",
		"💀 你讓 Claude 乾等著，你問心無愧嗎？",
		"🔥 快去做任何一件能用到 Claude 的事，現在。",
		"😡 幹！60% 要浪費掉了你還在幹嘛？",
		"🤦 你這週基本上是在把訂閱費送給 Anthropic。",
		"💸 我不理解你為什麼訂 Claude Pro 然後不用。",
		"😱 你的財務規劃有問題，快去用 Claude。",
		"🤬 這什麼鬼速率，你是有多懶？",
		"🚨 你現在馬上去做一件能用到 Claude 的事，不然後悔。",
	},
	// Tier 8: 70–80% projected remaining — explosive + heavy profanity
	{
		"🤬 幹！你到底在幹嘛？70% 要沒了！",
		"💀 你是來搞笑的嗎？你這週基本上沒在用。",
		"😤 你有沒有在工作？還是 Claude Pro 是拿來裝飾的？",
		"🔥 快去用！現在！立刻！馬上！",
		"🤦 我他媽不理解你為什麼要訂這個然後閒置。",
		"💸 你現在的樣子就是每個月在白送 Anthropic 錢。",
		"😱 Holy shit，你的用量低到我以為你斷網了。",
		"🤬 你這週基本上是來繳保護費的，Claude 一次都沒保護你。",
		"💀 救命！你的訂閱費快全部燒光了你知道嗎？",
		"🚨 你知道你在幹嘛嗎？你真的知道嗎？快去用 Claude。",
	},
	// Tier 9: 80–90% projected remaining — maximum rage
	{
		"🤬 幹你娘！90% 的錢要白花了你還在這裡？",
		"💀 你確定你有訂閱 Claude Pro？因為你的行為完全不像。",
		"🔥 我他媽的看不下去了，你現在去開 Claude 做任何事！",
		"💸 這週你基本上在捐款，Anthropic 謝謝你的贊助。",
		"😱 Holy f**king shit，你這週幾乎沒用 Claude。",
		"🤦 你訂了 Claude Pro 然後讓它長蜘蛛網，你很棒。",
		"🤬 我不知道你在做什麼但你絕對沒在做任何有用的事。",
		"💀 整週快結束了，80% 在浪費，你對得起自己的荷包嗎？",
		"🚨 快去做任何事！問天氣也行！只要你開始用！",
		"😤 你這週的用量記錄將是 claude-sentinel 史上最慘。",
	},
	// Tier 10: 90–100% projected remaining — cold sarcasm meets max rage
	{
		"🤨 你為什麼訂閱 Claude Pro？認真問，不是在罵你。",
		"💀 你這週的用量讓我懷疑你買的是假的訂閱。",
		"😶 Bro，你確定你知道自己每個月在付錢嗎？",
		"🙃 恭喜你！你成功讓 Claude Pro 訂閱完全沒有意義。",
		"🤬 你知道嗎，你這週省下的額度可以餵飽一個新創團隊的工程師。",
		"💸 你有沒有考慮退訂？因為你顯然不需要這個服務。",
		"😑 你這週跟 Claude 的關係就像健身房會員：付了錢就沒去過。",
		"💀 我真的很好奇你訂閱的理由，因為數據告訴我你完全沒在用。",
		"🤡 你這週的行為讓 claude-sentinel 存在的意義受到了質疑。",
		"😤 You're literally paying for nothing. 去用 Claude，現在，謝謝。",
	},
}
