package api_examples_test

import (
	"context"
	"fmt"
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	"github.com/iotaledger/iota.go/trinary"
	"net/http"
	"time"
)

const endpoint = "https://example-iri-node.io:14265"

var iotaAPI *api.API

func init() {
	var err error
	iotaAPI, err = api.ComposeAPI(api.HTTPClientSettings{URI: endpoint})
	if err != nil {
		panic(err)
	}
}

// i req: settings, The settings used for creating the Provider.
// i: createProvider, A function which creates a new Provider given the Settings.
// o: *API, The composed API object.
// o: error, Returned for invalid settings and internal errors.
func ExampleComposeAPI() {
	endpoint := "https://example-iri-node.io:14265"

	// a new API object using HTTP connecting to https://example-iri-node.io:14265.
	// this API object will use AttachToTangle() on the remote node.
	iotaAPI, err := api.ComposeAPI(api.HTTPClientSettings{URI: endpoint})
	if err != nil {
		// handle error
		return
	}

	// this API object will perform Proof-of-Work locally
	_, powFunc := pow.GetFastestProofOfWorkImpl()
	iotaAPI, err = api.ComposeAPI(api.HTTPClientSettings{
		URI:                  endpoint,
		LocalProofOfWorkFunc: powFunc,
	})
	if err != nil {
		// handle error
		return
	}

	// this API object will perform Proof-of-Work locally
	// and have a default timeout of 10 seconds
	httpClient := &http.Client{Timeout: time.Duration(10) * time.Second}
	iotaAPI, err = api.ComposeAPI(api.HTTPClientSettings{
		URI:                  endpoint,
		LocalProofOfWorkFunc: powFunc,
		Client:               httpClient,
	})
	if err != nil {
		// handle error
		return
	}

	_ = iotaAPI
}

// i req: uris, The URIs of the neighbors to add. Must be in udp:// or tcp:// format.
// o: int64, The actual amount of added neighbors to the connected node.
// o: error, Returned for API and internal errors.
func ExampleAddNeighbors() {
	iotaAPI, _ := api.ComposeAPI(api.HTTPClientSettings{URI: endpoint})
	added, err := iotaAPI.AddNeighbors("udp://iota.node:14600")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(added)
	// output: 1
}

// i req: trunkTxHash, The trunk transaction hash.
// i req: branchTxHash, The branch transaction hash.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: trytes, The transaction Trytes.
// o: []Trytes, The Trytes slice ready for broadcasting.
// o: error, Returned for invalid Trytes and internal errors.
func ExampleAttachToTangle() {
	bundleTrytes := []trinary.Trytes{
		"MINEIOTADOTCOM9MANUAL9PAYOUT99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999MFPZREFIGWHMM9YGSBSZBUBTKVUMNGOW9SST9YVHWKJMWSV9EZFSVPHIVNZQPLZEOTVOTSM9PLFKQOFRXXTC999999999999999999999999ENOEIOTADOTCOM9999999999999HZZGBYD99999999999C99999999ADFYHTLHGFJFQVVBN9CZO9QFPYOXAQFBJCC9VBHKWSKAKAZOGBOODFCCKZPIXZBIZTVFOUVJYWELRJUF9ROILTKTLTBFJBGOBVUSNNRPUQPCWXGZOVQMNYXOXMXRIXWMZOEUYZOKIFZUGLRQBAMUKNOZR9YKNZ9999UNRLCHFDXTONFKQGPVJB9KOFNPCEGKBXMDZVJWYHVX9VLSCYJNHWQWKCCUSHRGTDBFJRRHVTGDYOA9999MINEIOTADOTCOM9999999999999WLKYQANKE999999999L99999999YKUYYIHG9MYTHKHMB9BVGNPOSP9",
		"KXEGLNKUGRMKCCEAGRAAEZPEGOMNFI9QNIS9GMWJRZCHKAPL9BSFJNCM9AAOKFHLUHUJGZURESWDZADXWXUJBZQX9HFYKHCDVHQABEB9JAZANLYAKLOBJ9FBEM9VP9MITLYYPCLFWHTS9POOAKCP9TIJXPDYHSPGPAYMAMGG9USDJXOBTEXVPJRIQTZRRBMIEVU9FXIXOITLZHFEHUGZLMROJZODRNEFPK9DPVXQKJKGZT9KYLXZIODCQXDOFP9MQIXMYUIIZBFUALCOPMTV9CTOUXOKVKLKDAYVMBZCDFYTM9TEQFXBDMUIDLDCAOFBI9DWHNAIHP9KVPHMHFSXT9CQDL9XNJAXWQXPOVPDCXCJ9PQCJJCATAEBYZZ9YPSEYACXRJQUSANZUZQCJQJPWDOWHFQAIUUMPNFE9JZEAHKNCYQQHHIAKFETTNUXZXRM9XJFPAFAH9CWBDWVSVNAEABVBLGOVBMITV9TNXCVBGDWIEFFBNBFO9AGIXSCSMRFNJUDUMCKHQPYWFTKTPWRJNOHUGRRPAZB9LAP9PO9OOWCJWZOAWNTXVZEZAQHKPGEYABGOZOGTIHUANDLBXFHGALTPJNJWMVYTKHODJ9NVTHSVRRB9IIZGEXAPBTCNJKNCWMJPOCWNOZ9WZKHTGBTXMDCATAOLWZWAJEHCGZOWQRMDHXHEAAPMA9LKV9FUCZPHCNLHFIFMGVK9J9OAKEDCANF9PLPJRBPDRWBEOBJUGQLPSMALLPNKNIMQSPGAUMUZVCIXAMIVPCRLJLOOUVQMCQXQTQRSGWVORAMFGNULWYCWJVDXCCZNV9BOOHMWFV9BAGBDCSDTVNYWJT9RUWQFWOIHDHAG9EYWJETUHHXKWXOWI9JJJCUDZRBRDCBBEJKWGQVHK9FIURFUBZZQYGAMIXFFAIGUXBSYTKGDAEJKIZZU9JGEGX9JDIIZIMYIHJHRSSLDANHNQCCZYMGASJCDQJOICAQGZCDIKYMHLFQIKRNVOMDUUHUXBDVQDCOJRWNKYNMTBQFTWMJRHIK9KWIDMTNXGEN9YCQDJPSGVYYXDUYKKMPLPYOMTDVCCZUZZ9OFRDEAWKJ9HYHNQSBQNIWCFQI9RNSS9VD9ZQMRPAIQKIHOZWBNPCOLIKS9NRJCHGUCYPJHLENFBLEZAWOXZNOEVRNDHKGQOEGON9USZKVAGLXRJEOTPWEOTYPYENDAXASRPPYHLVNGONEIHQHTXIOANVTIFWKGBFURGFQXTCGXBEPRFTDZZLXFDQJFVILSJVL9LGSLXFDARPKICXDWQGUHQJTMFQVZNWX9NZYLL9LFXQWULHPFROJITRNHMRWEKSFKNQFKI9YDVKODHAPUAOOVATURGOPSQARXBYEIWZHITUQLPGWYZKATQOHQMYHBZYRYP9ZEXUFMRLIZLWHDRGXGLUFRESRNKDCCUFU9HNNYQH9VWVNCBVJHMIZWOUZJHZOQRZJNVQBMLKLMKDXGOSSMPHIEWMKXGFYZEYCOIZWFHPU9EJ9SRQWOTYXIVMMYXMVPUDREQDWJCIHDUQJKME9ZRGFBONBXXTRVFLVAOOWPYAORR9AELLDDFMQJWVJHMCHDWI9QLZZPLEQNDYGWCJWYJXWIMLXNUCJKTETTGCAWPXOIIOAGCODLEJTVTKVCMCDCF9IJLENMGIR9MLA9IFKLPQAIMHTAAHQSA9MYQWDKUSZSEWKHWJNXRMSLSWYELCCYVACBSPALFTEOOJETZOOZRPHBERRUWGDPMOGHJJWIWLMIDPMBLBIAWPTUSEMRHA9YKWGNVSQQUWNSSLKEMNIOLKWUUBHYKBAVAIAGXRYMGKEKGGMYRD9PKORVUHAEFCCPKUTPG9GASMFWQXKG9XEYAN9MAUWOMV9YILOVSRGAFNDFCP9LSWLX9QIOX9EYMCGQ9SGWMWITSPUAEQSMYHGXAX9TQMYOFKNBRHQKS9RJEPETEKLY9PBYNQBMQPRAQYCEBRFJCSGSTBSKTEDTNXSMNKYPL9FPGWILNHZ9UFXZIZQGFYKRUCDHCPGICHHJK9WUKVJ9AVABIMLZCXIJKAQDUSKGOSORG9OWVLFFBPQVMPIGTS9ABE99BMCYW99YYLSOMLVAXSZPMOBTGVZERVTZILSN9RHZAMC9WBUECMANLRJRLNJXITZFRNN9RS9GQBWIQBEYUY9TODJNLHUUVAOEPWOQQQMXXHFBSYZLGOUGXMGLFBGMHOLXQGDERVVPWHYDRPZCESCEXUMSO9RHVFZFDAUMKTWCICVO9999999999999999999999MINEIOTADOTCOM9999999999999HZZGBYD99A99999999C99999999ADFYHTLHGFJFQVVBN9CZO9QFPYOXAQFBJCC9VBHKWSKAKAZOGBOODFCCKZPIXZBIZTVFOUVJYWELRJUF99UDWPZIWOYEAFTOAIMPIJJCSHUFT9ZXS9HUUOXVRQFULFQLNSMUZLQJLELVY9BWFBBOEAPSGKCTL99999UNRLCHFDXTONFKQGPVJB9KOFNPCEGKBXMDZVJWYHVX9VLSCYJNHWQWKCCUSHRGTDBFJRRHVTGDYOA9999MINEIOTADOTCOM9999999999999S9KXQANKE999999999L99999999BOFHKXILIKCZFHKCVJGKYJSWCCW",
		"ZHTLLQHM9UHVTMQUTHVMMJP9YYXFGBQXBZALKEYQSQZOCUGMOPOH9ZFXZBJMQFZDDJYDKFLGAEVIWAHXYUHVHTO9DPSVWLUJZAQGGA9CSVQYXUNXBQPKHWGMRT9ARUEAFAACNIHWFPDMYQUDCXFPVSFBODGKXSRGXWMGFMTWCHYNLTZUWWLNGSEZXVWNXGLE9T9YQHLVUHTIERZNLKGPWSUEVXFVZIKUH9TEXUHTKUOTWG9QAADPFDEQONR9UWULKOTHUSTGMUBLWIGOKEQDSW9NAAKJDUOYKDIDMOVJMKSBQW9LQWTWW9BLRRNTYOTINPODAHSVJCSVGKZVFGSAPVWLTUOBTNLNWNJ9NCEDCWNIGNKUEYFWCOUJRKA9CEWBVD9PUSCUXUQDSTYWHU9SDWSGIAGP9TFJMFNCBDVLHDIIYZBYGUHKCBRHVZBRZXISJAHILKGRAIUBYOTB9NVHYZQTMSEAATFZKHBRUWCATZKDQWTOICUZLNIPXYPSSPWNULFASOFRAOLSMZSXDIFSVVY9JRBOVAYTJBTPAVJDSHZOVPJNVBDAJJYOHVIDZKCTZCZARIDDOXQGRQWLMWMGKNM9AJCJEKKHUIFCUAGALI9MMHKUPEUJYTYDVMQWURDXGURWPDUBNMTK9UUTNMRWTHINKWYKZIUGWBOZNAPOSAOQODGNN9FVT9OWNYTPLTUFBPIQCSWVMGSPEFPNEOJJSGVIAECMUQOO9LDHPXNXLBOVSYPPMNXHXLLFAZNKFRNFSMQWBH9YMJAX9FTINMXOOBGDIXHZNLQXBBBRBMSBBARSBUQPLCXDAUSKJFZEHSERAYIIGMOZYPKTNVILIFGE9WWWVTBRGUDYUOXEROFDCOTQAPMLNIZKRKJDFWWOXNXBOLAG9WYOYCKZYLIAIPNSVRHNOAAJNPTGUPMQRGDBGLCC9DQIDOTJ9RPA9XJKUSZHDEHJAWJRNQUXWFELDBXHFRLQZMTIQLGRKPGNAZWTDL9EAAUZWWOJJXSJNVMRZ9SX9GJWYPDAQUCJ9FELQWRSRQXHBZVWXINICZNQJFDZYCCLTRTBKBERUXNSGNYJWAXAEYXTKMRK9QXZEVURNHG9PCZHLOPPLSBLEFVZMETKJORKWWDTJHSUUWYVTA9MVDUVGQTHWYZRDVB9MRDDCFYAZ9XEB9OSDGMPMT9VCIYLZQNQJAMMBAPFECQUGG9MOYWFEL9JGLZSCUFVJ9XTNLSDNFUDLZRTHVCOMDNCSLGALDZDKKZNYLZRWOQGRNWXWGBCMYZNHSOHC9ERIQPANUR9RSVBWXYVHTERAD9BHMZETDCHQFSXLFXMVEGGHWRFTYEGXWDXOIMOKGBYHZSYKDXOMWUXNTTBJYCQBLBXHBQKXSLRMZCEDPKEZDENFNJUQNLSXWTKFSNWUGMYPFPWIQKVMEZTJ9NKMMMVYEDZPORRDYPHUJWSABIKOLEVEJWEJCQLLKEISRARYDQRA9QVWLRKQYLRELU9JIQUKFHUD9FTVBSJUCLWITWHXULDSXXKCILTJX9E9MRJWNZVKMQJOXHPBQPMGBLTCBNJP9CCEQQXBCVBNYYMOTKYTWZV9NPMBPRGSFODKX9AKGUKBHJEXWGCUFATN9QQEKMDOAATRNADL9FSGBLDCUKCNE9WUGLFIBPPMPYJYJVTUGICZUVNFPGUNXNNNXXZVQCQSBPVDODYLKGRAVJXIDRANTCUQLPVKOHBJQDTQYKCVMTGFOZVWHFMEOHYQCF99IIJW9WNKUVOVKBLOZBOIXNQIDHQLTXMZHLRYCAJBTAZJAD9NOPCFIRESBYJVVJAOGLGERIKS9VBVJ99OZHGZPHCIELKJWFENMUHTKOSULBZFAYNO9UXOCXAWARFUV9WWKNXL9YCRVIJ9ASNZHRPSELUBEMXWPD9ORFDBIPMEBEAAAIIFQUNWBNHDWYYJFMRYGKTPBIJLFFOJMTEBELGXRAGPRHYZDPGAIUOX9CMWCOBEVXAMGWT99RYMLICYWSVCBPEOHI9CGEEDGLPCG9EJPUJ9SNVZMTYXCHGVJAEZWWDSEQEWLH9BX9DOHDUZIWH9YBNYNRETUYRGGRXKEOQKGUQQMIQMCEKELKN9QDSJJNDXFIEYXTYW9NYKPZCQUGMWPIQUS9PFYCDHLGUSZFLBOJHRNRWXPGDEICJDXEYCPTNFECJNLHUUVAOEPWOQQQMXXHFBSYZLGOUGXMGLFBGMHOLXQGDERVVPWHYDRPZCESCEXUMSO9RHVFZFDAUMKTW999999999999999999999999999MINEIOTADOTCOM9999999999999HZZGBYD99B99999999C99999999ADFYHTLHGFJFQVVBN9CZO9QFPYOXAQFBJCC9VBHKWSKAKAZOGBOODFCCKZPIXZBIZTVFOUVJYWELRJUF9YD9HUVCQZLM9QEZKEOHDEYNSXJYAECXQURYAYGPWTAMFKPJIFTSE9QMREOQVDC9TDIL9KQKARIIE99999UNRLCHFDXTONFKQGPVJB9KOFNPCEGKBXMDZVJWYHVX9VLSCYJNHWQWKCCUSHRGTDBFJRRHVTGDYOA9999MINEIOTADOTCOM9999999999999UPIXQANKE999999999L99999999PRUXOLNLUASRN9PIOIU9BIOO9RM",
		"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999JPIGTDTEELOIEXKGCUDTVSNCYRFDSDXGYNIGZIPLSFGAMLWZCIHWFN9OFOTVUYQMTQYP9OTWHQGPXTBQW9YUEL9999999999999999999999MINEIOTADOTCOM9999999999999HZZGBYD99C99999999C99999999ADFYHTLHGFJFQVVBN9CZO9QFPYOXAQFBJCC9VBHKWSKAKAZOGBOODFCCKZPIXZBIZTVFOUVJYWELRJUF9UNRLCHFDXTONFKQGPVJB9KOFNPCEGKBXMDZVJWYHVX9VLSCYJNHWQWKCCUSHRGTDBFJRRHVTGDYOA9999YILLTIPHHTXRL9XJOVPNMLPEI9NXWUOXBIHOU9TVUGKYVGGIIWZODQ9IKEEJRTLDQQVNEWBZMXXO99999MINEIOTADOTCOM9999999999999JYEXQANKE999999999L99999999WKEBEQNMVNCLO9RMIKDLRO9MLOL",
	}

	tips, err := iotaAPI.GetTransactionsToApprove(3)
	if err != nil {
		// handle error
		return
	}

	finalTrytes, err := iotaAPI.AttachToTangle(tips.TrunkTransaction, tips.BranchTransaction, 14, bundleTrytes)
	if err != nil {
		// handle error
		return
	}
	_ = finalTrytes
}

// i req: trytes, The Trytes to broadcast.
// o: []Trytes, The broadcasted Trytes.
// o: error, Returned for invalid Trytes and internal errors.
func ExampleBroadcastTransactions() {
	// trytes which are chained together and had Proof-of-Work done on them
	var finalTrytes []trinary.Trytes
	_, err := iotaAPI.BroadcastTransactions(finalTrytes...)
	if err != nil {
		// handle error
		return
	}
}

// i req: hashes, The hashes of the transaction to check the consistency of.
// o: bool, Whether the transaction(s) are consistent.
// o: string, The info message supplied by IRI.
// o: error, Returned for invalid transaction hashes and internal errors.
func ExampleCheckConsistency() {
	txHash := "DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999"
	consistent, _, err := iotaAPI.CheckConsistency(txHash)
	if err != nil {
		// handle error
		return
	}

	fmt.Println("transaction consistent?", consistent)
}

// i req: query, The object defining the transactions to search for.
// o: Hashes, The Hashes of the query result.
// o: error, Returned for invalid query objects and internal errors.
func ExampleFindTransactions() {
	txHashes, err := iotaAPI.FindTransactionObjects(api.FindTransactionsQuery{
		Approvees: []trinary.Trytes{
			"DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999",
		},
	})
	if err != nil {
		// handle error
		return
	}
	fmt.Println(txHashes)
}

// i req: addresses, The addresses of which to get the balances of.
// i req: threshold, The threshold of the query, must be less than or equal 100.
// i: tips, List of hashes, if present calculate the balance of addresses from the PoV of these transactions.
// o: *Balances, The object describing the result of the balance query.
// o: error, Returned for invalid addresses and internal errors.
func ExampleGetBalances() {
	balances, err := iotaAPI.GetBalances(trinary.Hashes{"LWVVGCWMYKZGMBE9GOCB9J9QALRKWGAVISAEXEOM9NVCGJCCGSNBXXGYQDNZBXBWCEM9RMFHYBCSFWE9XEHAPSXHRY"}, 100)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(balances.Balances[0])
}

// i req: txHashes, The transaction hashes to check for inclusion state.
// i req: tips, The reference tips of which to check whether the transactions were included in or not.
// o: []bool, The inclusion states in the same order as the passed in transaction hashes.
// o: error, Returned for invalid transaction/tip hashes and internal errors.
func ExampleGetInclusionStates() {
	txHash := "DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999"
	info, err := iotaAPI.GetNodeInfo()
	if err != nil {
		// handle error
		return
	}
	states, err := iotaAPI.GetInclusionStates(trinary.Hashes{txHash}, info.LatestMilestone)
	if err != nil {
		// handle error
		return
	}

	fmt.Println("inclusion?", states[0])
}

// o: Neighbors, The Neighbors of the connected node.
// o: error, Returned for internal errors.
func ExampleGetNeighbors() {
	neighbors, err := iotaAPI.GetNeighbors()
	if err != nil {
		// handle error
		return
	}
	fmt.Println("address of neighbor 1:", neighbors[0].Address)
}

// o: *GetNodeInfoResponse, The node info object describing the response.
// o: error, Returned for internal errors.
func ExampleGetNodeInfo() {
	nodeInfo, err := iotaAPI.GetNodeInfo()
	if err != nil {
		// handle error
		return
	}
	fmt.Println("latest milestone index:", nodeInfo.LatestMilestoneIndex)
}

// o: Hashes, A set of transaction hashes of tips as seen by the connected node.
// o: error, Returned for internal errors.
func ExampleGetTips() {
	tips, err := iotaAPI.GetTips()
	if err != nil {
		// handle error
		return
	}
	fmt.Println(tips)
}

// i req: depth, How many milestones back to begin the Random Walk from.
// i: reference, A hash of a transaction which should be approved by the returned tips.
// o: *TransactionsToApprove, Trunk and branch transaction hashes selected by the Random Walk.
// o: error, Returned for internal errors.
func ExampleGetTransactionsToApprove() {
	tips, err := iotaAPI.GetTransactionsToApprove(3)
	if err != nil {
		// handle error
		return
	}
	fmt.Println("trunk", tips.TrunkTransaction)
	fmt.Println("branch", tips.BranchTransaction)
}

// i req: hashes, The hashes of the transactions of which to get the Trytes of.
// o: []Trytes, The Trytes of the requested transactions.
// o: error, Returned for internal errors.
func ExampleGetTrytes() {
	trytes, err := iotaAPI.GetTrytes("CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(trytes[0])
	// output: ZYVOTKSGSVLVAHEBG9TQJVBASNIFUXGJVZ9YIQFFJAXLRQDRKA9RXFIDVHONTBJPZ9UCKPGZQQN9BDAKCJCLHR9ZNQTHRYR9GVZXT9JJFLIORBTTXEOWKGKWUVREIFGF9NVYSCUKFCSMAERHGCYDNYFX9SYBMDVXJ9QE9MCZLDFBOOASQELBVCQKGKDBEYGFOOJKNGZ9CICUBHZMCGJGGLVXIKEDDWUCWDCMPMIMNVBXYYNQISWNRGZRKSEFHB9USXTIEWOYXDPDFAKFEMS99NTDZRSAJSBSOCKWVVJIOHHENIGLNNSPFUAPNCDULETAQNNDOF9FGFYTOSMFYSKLLPLBDRASEDUJOBBEETIIDD9BJWF9MLWZNXN9XZSCJZHQWFVYPBGMNKDHIXZHLFIEXL9FPUUX9KARDTNGVDZNBPCBNTSGKUUFAJDZGPGTWRKXEYRBYFPAZPKZ9XZLVUKTYUPKXUBCSTHJPABBSZJNJDVYMLRDGBGQWCANSZIJCQAPL9LBVHTHHFXFUDZKKFV9O9IWKYPK9JQVLGXVQPUCFFFTEHLDEIROIQWKEKVMPAUTAHVUQXHXRIDQNGYVFLPZHOBAQFIGEIIZSRWLETZVWVLOYF9JTEVPEMJHMXAUHQSEOOBOMJYAEGBSASSF9BHWCOCUOYTKH99RISXDKIACLXNEFWKYJD9U9KPCFSMFGYDMLFKJNIMTKJPJGZLKUZVHLYTWACYTCLPSHXARNMMGEZQLVAYJOSJIVNXJRSPRFHVPIHULYBXEWHIUYOGZNWCRIUSZRYGGZALGDZNZXA99BD9IQEWEBXALRAVOUKDAMBENLTTERRNJEEPYLG9V9MABZOWWFTMFMRWWUQIFROCCUBOKABIIPDCRBP9QEHZRHRASPBEALUBPIRTCMYRIKZTTCLYCLVOJEZEFLYRXABVPZWCKWJJJEQLWCDBMLSQTEJRPJITMJHBAAKD9SSKQHTFZWEOAVDERPLOTOK9EXJBXSZABZUCMCTEJY9P9CZKYLMXRZBLX9PRKVRUMVURTVTUDJUEWPDDBCVCCPIFAUOXOK9AAMFDSPKEGOMHSVESUUEFIPSGQTEUWDHTAPRTOGEPVNFWIHMSJVDEKCOKV9PWDQNFAYXQJHVABMVVNNISXPQYQHSIDFYOMDAHRAPCZRQOQONXSRDDUYKHF9WKBWXXSQWKZKWP9NEEKNXPOQLETRUWXMMRZ9SHTO9IAHKUOMXI9EMUEMNAKXZETHMXGMIIPZQY9WXYFHBKMLVHRSEPZVNIHPUNPXQCKAKWYTOAASXNCAHGCOTUGSHYURUJXPRBF9RMHQB9AKBUEZNTYUFPIQQKYBLMDKPMVV9CRUQGOIWCWJPNUYFYZ9FO9DKWDEKD9ELAEFCXQUOOZGPVRXBGNAHQSMUZRUENYKFNWWUATSASTKKP9HNA9LCJEBKFUOAMBKKDGSP9WAWJWAEUTF9MIEBF9MFOCYZYAGSN9VDOWUYD9SHIYBVSG9CJQJSLCVMSAUJFZCCLEDHI9IXGDNILDWAJRRAAJNAGCSARCRQBXKARTSZCALDSAXXZCXSEGWWAOYCMQU9L9KYJCHSJJAEBXJZAORMXISBMKJYEZEMGKVMB9YZSAXMEQUHHABSXROERMLBHKDAHOSVRGJWDDZTDKRAJUWSBSNVASMFVHNDSYBEPZOSCMZQNQRVLVUJKKCESMJGMKYNEDYNDENMSZRGPYZFXFQLKVTYTUZJSUYZBBPGKIUDJFCXCSYXJJFLVEYUDBVISZUNINDTXUJXNCYSXHLGNIWMZEZPAVSU9YDBMSVIGKRHAMS9UPLIZANLFUESVTRFVWP9SPRPD9F9ILJIXMFRWKZYCZUIDNPNGFUJLOEQWFUJROZTOILCC9NYZAMKNZKIFQBIEJZCYPOVDGUBPSETFRXVTIDKCTY9TBBRUPFMXMNHY9REKOYVAAZFFGC9D9BPSVRYKNRVUMLNPVYS9FNSPGJMYLSPNYSKJSLEIJPJSPZJDRDYUEWKRJBOOZKAODYUQEBZRRLEOGUDTSIELWKCDNTVQJCGWIRZCNJXYSSDG9ELNRNGDBAVCFFWSAJWZKMRBWVCOFROYTDNKYJBRYUR9KDMPU9JDATLLIFXULLZOAIKWZOZYWIKSYCNUHAZKCRDIKCJQDWADJTYJYRKXES9UKNDFOFUGZHOAAETRKPXQNEK9GWM9ILSODEOZEFDDROCNKYQLWBDHWAEQJIGMSOJSETHNAMZOWDIVVMYPOPSFJRZYMDNRDBSCZ99999999999999999999999GC9999999999999999999999999CBDAMZD99999999999I999999999RVBBAEIONQBWLFIOLFQTEETLYCULDNPK9LIUAE9EIDUU9VDVL9OD9LHLEPYFEDJPFHOSTZCBILVKCRTYXOOGSSQFKDLHWUNQGIWJSYAL9AUMKINSYUKTMZBZKXVETCJLPX9D9KWPBZYUYFTFWOFFLIVKLCFDA9999FAPCHJMEDCUAIYOILWUTGFNTI9UWKTQKNMUEFQCNICFZNUSKPVIXVKSEIPSGEXHIKMWYCORQXYGD99999GC9999999999999999999999999SSNZBWLLE999999999MMMMMMMMMQFAVPOJVPZDLNJUHF9HSFWJUQDD
}

// o: error, Returned for internal errors.
func ExampleInterruptAttachToTangle() {}

// i req: uris, The neighbors to remove.
// o: int64, The amount of neighbors which got removed.
// o: error, Returned for internal errors.
func ExampleRemoveNeighbors() {
	removed, err := iotaAPI.RemoveNeighbors("udp://example-iri.io:14600")
	if err != nil {
		// handle error
		return
	}
	fmt.Println("neighbors removed:", removed)
}

// i req: trytes, The transaction Trytes to store.
// o: []Trytes, The stored transaction Trytes.
// o: error, Returned for internal errors.
func ExampleStoreTransactions() {}

// i req: addresses, The addresses to check for spent state.
// o: []bool, The spent states of the addresses.
// o: error, Returned for internal errors.
func ExampleWereAddressesSpentFrom() {
	spentStates, err := iotaAPI.WereAddressesSpentFrom("LWVVGCWMYKZGMBE9GOCB9J9QALRKWGAVISAEXEOM9NVCGJCCGSNBXXGYQDNZBXBWCEM9RMFHYBCSFWE9XEHAPSXHRY")
	if err != nil {
		// handle error
		return
	}
	fmt.Println("address spent?", spentStates[0])
}

// i req: tailTxHash, The hash of the tail transaction of the bundle.
// o: []Trytes, The Trytes of all transactions of the bundle.
// o: error, Returned for invalid tail transaction hashes and internal error.
func ExampleBroadcastBundle() {
	hash := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	bundleTrytes, err := iotaAPI.BroadcastBundle(hash)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(bundleTrytes)
}

// i req: seed, The seed of the account.
// i req: options, Options used for gathering the account data.
// o: *AccountData, An object describing the current state of the account.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetAccountData() {
	seed := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	accountData, err := iotaAPI.GetAccountData(seed, api.GetAccountDataOptions{})
	if err != nil {
		// handle error
		return
	}
	fmt.Println(accountData)
}

// i req: tailTxHash, The hash of the tail transaction of the bundle.
// o: Bundle, The Bundle of the given tail transaction.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetBundle() {
	hash := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	bundle, err := iotaAPI.GetBundle(hash)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(bundle)
}

// i req: addresses, The addresses of which to get the bundles of.
// i: inclusionState, Whether to set the persistence field on the transactions.
// o: Bundles, The bundles gathered of the given addresses.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetBundlesFromAddresses() {
	addresses := trinary.Hashes{
		"PDEUDPV9GACEBLYZCQOMLMHOQWTBBMVMMYUDKJKVFVSLMUIXHUISQGFJKJABIMAVRNGOURDQBBRSCTWBCNYMIBWIZZ",
		"CUCCO99XUKMXHJQNGPZXGQOTWMACGCQHWPGKTCMC9IPOXTXNFTCDDXTUDXLOMDLSCRXKKLVMJSBSCTE9XRCB9FGUXX",
	}
	bundles, err := iotaAPI.GetBundlesFromAddresses(addresses)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(bundles)
}

// i req: txHashes, The hashes of the transactions to check for inclusion state.
// o: []bool, The inclusion states.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetLatestInclusion() {
	txHash := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	inclusionStates, err := iotaAPI.GetLatestInclusion(trinary.Hashes{txHash})
	if err != nil {
		// handle error
		return
	}
	fmt.Println("inclusion state by latest milestone?", inclusionStates[0])
}

// i req: seed, The seed from which to compute the addresses.
// i req: options, Options used during address generation.
// o: Hashes, The generated address(es).
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetNewAddress() {
	seed := "HZVEINVKVIKGFRAWRTRXWD9JLIYLCQNCXZRBLDETPIQGKZJRYKZXLTV9JNUVBIAHAGUZVIQWIAWDZ9ACW"
	addr, err := iotaAPI.GetNewAddress(seed, api.GetNewAddressOptions{Index: 0})
	if err != nil {
		// handle error
		return
	}
	fmt.Println(addr)
	// output: PERXVBEYBJFPNEVPJNTCLWTDVOTEFWVGKVHTGKEOYRTZWYTPXGJJGZZZ9MQMHUNYDKDNUIBWINWB9JQLD
}

// i req: address, The address to check for used state.
// o: bool, Whether the address is used.
// o: error, Returned for invalid parameters and errors.
func ExampleIsAddressUsed() {}

// i req: hashes, The hashes of the transaction to get.
// o: Transactions, The Transactions of the given hashes.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetTransactionObjects() {}

// i req: query, The object defining the transactions to search for.
// o: Transactions, The Transactions of the query result.
// o: error, Returned for invalid parameters and internal errors.
func ExampleFindTransactionObjects() {}

// i req: seed, The seed from which to derive the addresses of.
// i req: options, The options used for getting the Inputs.
// o: *Inputs, The Inputs gathered of the given seed.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetInputs() {}

// i req: addresses, The addresses to convert.
// i req: balances, The balances of the addresses,
// i req: start, The start index of the addresses.
// i req: secLvl, The used security level for generating the addresses.
// o: Inputs, The computed Inputs from the given addresses.
func ExampleGetInputObjects() {}

// i req: seed, The seed from which to derive the addresses of.
// i req: options, Options for addresses generation.
// o: Bundles, The Bundles gathered from the given addresses.
// o: error, Returned for invalid parameters and internal errors.
func ExampleGetTransfers() {}

// i req: tailTxHash, The tail transaction to check.
// o: bool, Whether the transaction is promotable.
// o: error, Returned for invalid parameters and internal errors.
func ExampleIsPromotable() {}

// i req: seed, The seed from which to derive private keys and addresses of.
// i req: transfers, The transfers to prepare.
// i req: options, Options used for preparing the transfers.
// o: []Trytes, The prepared Trytes, ready for Proof-of-Work.
// o: error, Returned for invalid parameters and internal errors.
func ExamplePrepareTransfers() {
	seed := "IAFPAIDFNBRE..."

	// create a transfer to the given recipient address
	// optionally define a message and tag
	transfers := bundle.Transfers{
		{
			// must be 90 trytes long (inlcude the checksum)
			Address: "ASDEF...",
			Value:   80,
		},
	}

	// create inputs for the transfer
	inputs := []api.Input{
		{
			// must be 90 trytes long (inlcude the checksum)
			Address:  "BCEDFA...",
			Security: consts.SecurityLevelMedium,
			KeyIndex: 0,
			Balance:  100,
		},
	}

	// create an address for the remainder.
	// in this case we will have 20 iotas as the remainder, since we spend 100 from our input
	// address and only send 80 to the recipient.
	remainderAddress, err := address.GenerateAddress(seed, 1, consts.SecurityLevelMedium, true)
	if err != nil {
		// handle error
		return
	}

	// we don't need to set the security level or timestamp in the options because we supply
	// the input and remainder addresses.
	prepTransferOpts := api.PrepareTransfersOptions{Inputs: inputs, RemainderAddress: &remainderAddress}

	// prepare the transfer by creating a bundle with the given transfers and inputs.
	// the result are trytes ready for PoW.
	trytes, err := iotaAPI.PrepareTransfers(seed, transfers, prepTransferOpts)
	if err != nil {
		// handle error
		return
	}

	fmt.Println(trytes)
}

// i req: seed, The seed from which to derive private keys and addresses of.
// i req: depth, The depth used in GetTransactionsToApprove().
// i req: mwm, The minimum weight magnitude to fufill.
// i req: transfers, The transfers to prepare and send off.
// i req: options, The options used for preparing and sending of the bundle.
// o: Bundle, The sent of Bundle.
// o: error, Returned for invalid parameters and internal errors.
func ExampleSendTransfer() {}

// i req: tailTxHash, The hash of the tail transaction.
// i req: depth, The depth used in GetTransactionsToApprove().
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: spamTransfers, The spam transaction used for promoting the given tail transaction.
// i req: options, Options used during promotion.
// o: Transactions, The promote transactions.
// o: error, Returned for inconsistent tail transactions, invalid parameters and internal errors.
func ExamplePromoteTransaction() {
	tailTxHash := "SLFKTBMXWQPWF..."

	promotionTransfers := bundle.Transfers{bundle.EmptyTransfer}

	// this will create one promotion transaction
	promotionTx, err := iotaAPI.PromoteTransaction(tailTxHash, 3, 14, promotionTransfers, api.PromoteTransactionOptions{})
	if err != nil {
		// handle error
		return
	}

	fmt.Println("promoted tx with new tx:", promotionTx[0].Hash)

	// options for promotion
	delay := time.Duration(5) * time.Second
	// stop promotion after one minute
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*time.Minute)
	opts := api.PromoteTransactionOptions{
		Ctx: ctx,
		// wait for 5 seconds before each promotion
		Delay: &delay,
	}

	// this promotion will stop until the passed in Context is done
	promotionTx, err = iotaAPI.PromoteTransaction(tailTxHash, 3, 14, promotionTransfers, opts)
	if err != nil {
		// handle error
		return
	}
	fmt.Println("promoted tx with new tx:", promotionTx[0].Hash)
}

// i req: tailTxHash, The hash of the tail transaction of the bundle to reattach.
// i req: depth, The depth used in GetTransactionstoApprove().
// i req: mwm, The minimum weight magnitude to fulfill.
// i: reference, The optional reference to use in GetTransactionsToApprove().
// o: Bundle, The newly attached Bundle.
// o: error, Returned for invalid parameters and internal errors.
func ExampleReplayBundle() {}

// i req: trytes, The transaction Trytes to send.
// i req: depth, The depth to use in GetTransactionsToApprove().
// i req: mwm, The minimum weight magnitude to fulfill.
// i: reference, The optional reference to use in GetTransactionsToApprove().
func ExampleSendTrytes() {}

// i req: trytes, The Trytes to store and broadcast.
// o: []Trytes, The stored and broadcasted Trytes.
// o: error, Returned for invalid parameters and internal errors.
func ExampleStoreAndBroadcast() {}

// i req: trunkTxHash, The hash of the tail transaction of the bundle.
// i req: bndl, An empty Bundle in which transactions are added.
// o: Bundle, The constructed Bundle by traversing through the trunk transactions.
// o: error, Returned for invalid parameters and internal errors.
func ExampleTraverseBundle() {}

// i req: settings, The Settings used for creating a new HTTP Provider.
// o: Provider, The created Provider.
// o: error, Returned for invalid settings.
func ExampleNewHTTPClient() {}
