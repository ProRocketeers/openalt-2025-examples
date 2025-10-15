# slide 1 - 0:10
Dobrý den, jmenuju se Lukáš Stuchlík a rád bych vám tady něco řekl o alternativě k Helmu, aneb jak se nezbláznit z YAML templatů
# slide 2 - 0:43 (0:53)
V kostce o mně, vystudoval jsem magisterské studium na Ostravské univerzitě v oboru Informační systémy, a od té doby dělám už 5 let pro ostravskou firmu ProRocketeers. Jako firma děláme agilní vývoj softwaru pro klienty, kde převážně děláme Team as a Service. Já osobně se věnuju částečně nějakému backend vývoji v Typescriptu a Go, ale převážně jsem jako DevOps Engineer - správa CI/CD pipelines, Kubernetes clusterů, dělám s Helmem, Dockerem, Ansible apod. Ve volném čase se rád věnuju tomu přispět něco do opensource komunity zpátky, mimo jiné i do projektu, který vám tady budu představovat

# slide 3 - 0:34 (1:28)
Téma, kterým se budu s vámi dneska zabývat, je správa aplikací v Kube clusterech. Ukočírovat všechny svoje deploynuté aplikace spolu se všemi závislostmi a mít v nich přehled a kontrolu ve dnešní době už není taková bolest jako dříve. Veliká většina opensource (anebo aspoň self hosted) aplikací bývá distribuovaná i ve formě Helm chartů, a deploynout je je poměrně jednoduché, i v rámci GitOps systémů jako třeba ArgoCD. Podívejme se nejprve trochu více zblízka na Helm.

# slide 4 - 1:46 (3:15)
Helm je package manažer pro Kubernetes. Jeho ústředním konstruktem jsou tzv. Charty, což je obvykle pouze kolekce YAML šablon, které krmíme parametry (Helm values), někdy i s přidanými závislostvmi ve formě jiných sub-chartů. Když pak zavoláme Helm, aby nám takový Chart nasadil, tak posbírá values z různých zdrojů - obvykle to bývají nějaké default hodnoty přímo uvnitř Chartu, a pak různé vaše overridy. Typicky se takhle předávají celé soubory (tzv. values fily), ale Helm si rád poradí i s CLI argumenty, kam se dají protlačit třeba env proměnné apod.

Tyto parametry pak spolu s YAML šablonami jdou přes Helm templating engine, který je založený na jazyce Go, a prakticky takhle dosadí ty vaše parametry na příslušné místa v těch šablonách a výsledkem z toho jsou Kube objekty v YAMl formátu. Zní vcelku jednoduše, že? A ono tomu tak ve většině případů i bývá, a většina chartů se dá s trochu přimhouřenýma očima i docela slušně číst. Tím, že to není jenom pouhá substituce, ale jde to přes ten template engine, tak v těch šablonách můžete dělat i všelijakou logiku. Častokrát se takhle abstrahují třeba různé společné části kódu, nebo se provádějí nějaké manipulace se stringy, dávají se dohromady různé způsoby jak nakonfigurovat totéž, umí to základní flow control kódu, proměnné atd. Tento template engine vám dovolí dělat spoustu věcí. A to se vám může někdy vymstít.

# slide 5 - 2:07 (5:22)
U nás v ProRocketeers míváme i různé malé služby a projekty, které jsou pod naší kontrolou a potřebují taky nasadit na některé z našich vlastních clusterů. Dříve jsme to řešili typicky tak, že jsme používali předgenerované Helm charty co vám vyhodí příkaz "helm create", které jsme si upravovali dle potřeb daného projektu. Tento přístup ale byl strašně repetitivní, a když jsme se rozhodli dělat nějaké změny globálně napříč všemi aplikacemi, dalo to vždycky hroznou spoustu práce to upravit všude. Vzal jsem si tedy na sebe nelehký úkol, a dal se do přípravy vlastního "obecného" Helm chartu, který bude centralizovaný a společný pro všechny takovéhle projekty, a který obsáhne všechno, co tak může běžná aplikace potřebovat. A že jich ve výsledku není úplně málo, mezi běžně používané věci můžu vyjmenovat třeba velkou konfiguraci Deploymentu samotného (porty, envy z různých zdrojů, volumy, init kontejnery, sidecar kontejnery, resourcy), ale i věci okolo, např. Ingress, ExternalSecrety, HorizontalPodAutoscaler, RBAC práva pro ServiceAccounty a plno jiných věcí.

Vzhledem k tomu, že jsem k tomu chtěl mít jaksi.. hezký formát těch values, aby i developer experience (tedy naše) s používáním tohoto chartu byla přívětivá, nevyhnutelně přišla na řadu dost velké množství logiky. Různý preprocessing těchto values, validace, upravení těch vstupních hodnot do formátu, ve kterém se zase tomu template enginu s tím lépe pracuje atd. Kdo tohleto někdy dělal, začíná možná chápat, na co tady nejvíc narážím. Jde o to, že ačkoliv vám to Helm umožní, je to naprostá katastrofa s tím pracovat a udržovat to.

# slide 6 + 7 - 3:00 (8:23)
Podpora IDE pro Helm templaty není moc slavná, častokrát jste rádi alespoň za nějaký syntax highlighting. Buď jsem měl smůlu nebo neumím hledat, ale nepodařilo se mi najít extension pro VSCode, který by mi usnadnil práci podobně jako klasické programovací jazyky, a to i na takové banální věci jako IntelliSense, nebo nějaké napovídání. Takže když jsem se pak začal potýkat s poměrně složitýma datovýma strukturama, musel jsem se spoléhat na to, že neudělám nějakou chybu, nepopletu si úroveň zanoření atd. S tímto ještě půl bídy, chybové hlášky v případě že se snažíte dereferencovat proměnnou, která třeba neexistuje, tak tam vám Helm alespoň řekne nějaké vodítka, třeba název toho co se snažíte přečíst.

Horší problém pramení z toho, že těmi šablonami vytváříte YAML soubory, a ty jsou závislé na odsazení - whitespace. Helm templaty, resp. ty dynamické části, které píšete do dvou složených závorek, totiž mají na každé straně optional pomlčku, a ta tomu enginu říká, že z téhle strany "spolkni" veškerý whitespace až do prvního normálního znaku. A řeknu vám, když koukáte do stovek řádků takovýchto template, tak je *strašně jednoduché* někde takovou pomlčku buď zapomenout, anebo ji dát někam, kde být nemá. A nikdo vás na to neupozorní, že je něco špatně až na to, že vám začne render padat na YAML syntax chybách, ze kterých nejde poznat, kde je co špatně. Je super, že vám to řekne řádek a sloupec toho, kde je nějaká chyba, ale častokrát žádný takový řádek ani neexistuje. Nezbývá vám pak nezbývá nic moc jiného než pozorně pročítat řádek po řádku a hledat, kde je zdroj takové chyby. 

Další problém - řekněme, že si chcete ušetřit budoucí nervy, a rozhodnete se zavést nějaké testování, ideálně unit testování do svého Chartu. Ne že by to nešlo, existuje na to Helm plugin - `helm-unittest`, kdy i ty samotné unit testy píšete v YAMLech, ale ani práce s ním není úplně bezchybná (např. jsem se dostával do nekonzistencí, že lokálně při vývoji mi testy procházely, ale v CI/CD pipeline už ne a nepřišel jsem na to proč). 

Dlouho mi tyto problémy ležely v žaludku a chtěl jsem něco, čím bych to nahradil, a to mě přivádí k našemu konkurentovi, kterého vám chci představit - Yoke.

--- 
# slide 8 1:39 (10:03)
Yoke se představuje taky jako package manažer, ale *jinak*. Místo textových YAML šablon se rozhodl jít jinou cestou, a tyhle Kube objekty (často nazývané resourcy) vytváří v klasickém programovacím jazyce, čímž se o něco víc blíží klasickému pojmu "infrastructure as code". Pojďme si to trochu přiblížit. Jelikož je to pointou celé mojí prezentace, rovnou to budu dávat do srovnání s Helmem, který je lidem většinou známý.

Začněme u Chartů - jejich ekvivalent v Yoku se jmenuje Flight. Yoke celkově bere svoje názvosloví z aviatiky - netuším moc proč, možná má autor prostě rád letadla? Kdoví. Flight je obecně řečeno program v libovolném jazyce, který je zkompilovaný do WebAssembly formátu, a který na vstupu bere libovolné parametry, a na výstupu dává Kubernetes resourcy. Proč zrovna WebAssembly? Z několika důvodů: 

1) Je to společný kompilační cíl pro více různých jazyků - např. Go, C, C++, Rust, ale postupně i další - Swift, Zig, Python, .NET ....
2) Je nezávislý od cílového OS nebo architektury, protože produkuje neutrální bytekód, podobně jako Java
3) Jeho runtime je sandboxed, tj. nemá přístup k souborovému systému anebo síti

# slide 9 + 10 3:42 (13:45)
I když můžete teoreticky psát Flighty v různých jazycích, je tak nějak obecně doporučované vybrat si právě Go, a to z pár důvodů - jednak Yoke distribuuje sadu pomocných SDK právě pro Go, ale hlavní důvod za mě je oficiální podpora Kubernetes API typů, které si importujete do vašeho Flightu jako balíčky. Tím rovnou narážím na (z mého pohledu) hlavní výhodu, a to je plná podpora programovacího jazyka vaší volby. Můžete si svůj Flight napsat jako klasický Go program, používat všechny featury tohoto jazyka (i to co vám Helm v templatách nedovolil), máte k tomu plnou podporu IDE, klasickou type safety (a to nejen ve vašich datových strukturách ale právě i v Kube API strukturách), dokonce i ten unit testing. 

Je to možná trochu nezvyk nad tím takhle přemýšlet, a určitě je to o něco méně čitelné v porovnání s jednoduššími Helm charty (protože Go je v tomhle prostě ukecané), ale jakmile se už musíte potýkat s něčím složitějším (viz ta monstrozita, kterou jsem zmiňoval dříve), pracovat s tím jako s klasickou kompilovatelnou aplikací byla hrozná jako... úleva. Najednou můžu jednoduše refactorovat, dělat změny aniž bych musel těžce přemýšlet kde všude se to projeví, protože jednak na mě začne řvát Go, jednak mi začnou padat tentokrát už normální testy. Dá se tam dokonce debugovat klasickým způsobem, že si nastavím v tom programu breakpointy a můžu ten Flight krokovat! Tím, že Flighty jsou vlastně klasické programy, dokonce se dá vytvořit něco jako knihovna takových malých "stavebních kostek", třeba funkce, nebo nějaké reusable komponenty, které můžete distribuovat jako normální balíčky. Já jsem asi dost zaujatý, ale tohle všechno dohromady mi přijde jako fantastická změna k lepšímu oproti tomu jak jsem to měl s Helmem.

Nicméně možnosti Yoku nekončí tady. Dejme tomu, že jste v situaci, kdy se snažíte takhle zkonvertovat nebo přepsat váš Helm chart, který ovšem si nese *závislosti* na dalších sub-chartech (např. Postgres databázi nebo Redis), a ty jako nejsou vaše, takže nedává smysl abyste přepisovali i je. Yoke na tyhle situace reaguje tak, že má vcelku slušnou kompatibilitu s Helmem. 
# TODO: aspoň odrážky, je tu dost keců bez vizuální podpory 
1) Můžete zabalit existující Helm chart, a renderovat ho přímo z Yoke Flightu (pomocí Go direktivy `go:embed`, přičemž se z toho archivu stane pole bajtů v paměti)
2) Anebo můžete vyzkoušet automaticky *vygenerovat* Yoke Flight z Helm Chartu jako takového - `helm2go`
  - tenhle CLI tool opět zabalí existující Helm chart, ale navíc k tomu zkusí z JSON schématu Chartu vygenerovat Go struktury, které to reprezentují, čehož můžete taky využít dle libosti
  
# slide 11 - 0:31 (14:16)
Zatím jsme se tady bavili hlavně o tom, co Yoke dokáže, ale pojďme se podívat teďka na to, jakým způsobem oba tooly (Helm i Yoke) fungují ve světě GitOps, specificky právě ArgoCD.

Helm je vysloveně nativně integrovaný do Arga, bez jakékoliv konfigurace. Můžete tam specifikovat vstupní parametry, číst je ze souborů, ale taky inline v Application manifestech. Charty můžete používat lokálně z připojeného Git repozitáře, nebo z Helm repozitáře, nebo jako OCI artefakt. Argo Helm využívá ovšem pouze na ten *templating*, čili nemá to k dispozici např. funkci `lookup()` kterou Helm umí číst živý stav z Kube clusteru (což za mě je vcelku dobrá vlastnost, je to trochu bezpečnější). 

Oproti tomu Yoke je v tomhle světě něco trochu cizího. Naštěstí ArgoCD má koncept CMP - Content Management Plugin. Je to způsob, jakým si můžete integrovat libovolný způsob, kterým renderujete Kube resourcy a o zbytek se už Argo postará jako dříve. Z hlediska DX je to pro nás trochu nepohodlnější, ale bohužel s tím už nic neuděláme, to má pod taktovkou ArgoCD a ne Yoke. Dobrá zpráva je, že i s omezeními, které nám diktuje CMP specifikace, jsme stále schopni dosáhnout prakticky stejného výsledku jako Helm.

# slide 12 - 1:38 (15:55)
Technická realizace Yoke pluginu pro ArgoCD (nazývaného YokeCD) je takováto:
- ArgoCD samotné se deployuje do clusteru jako asi 6 komponent, z čehož jedna je "repo-server"
- CMP specka diktuje, že pluginy se přidávají do Arga jako sidecar kontejnery vedle hlavního "repo-server" kontejneru, přičemž použijou společnou binárku pro start CMP serveru + Plugin specifikaci, kterou Argo umí přečíst
- YokeCD je ovšem ve své implementaci trochu složitější, a tak se do Podu přidává další kontejner ("yokecd-svr"), což je jednoduchý HTTP server, který komunikuje s YokeCD plugin kontejnerem
    - tato spolupráce spočívá v tom, že yokecd plugin kontejner působí jako prostředník, jakási fasáda mezi repo serverem (se kterým komunikuje přes gRPC), a tím yokecd serverem (přes HTTP)
    - yokecd plugin pak od repo serveru přijímá input parametry, které různě zpracuje a pošle dál na server kontejner
    - ten se postará o stažení a caching našeho Flightu, jeho samotné spuštění a tím render těch Kube resource, které pak zpětně vrátí až do repo serveru

Je to trochu komplikovanější než by bylo libo, ale výhodou je to, že i navzdory tomu, že Flight je i v zabaleném formátu třeba 15MB soubor, tak je render díky cachování stále velmi rychlý.

Yoke má oproti Helmu navíc ještě jednu zajímavou feature, ale pro její představení bude lepší, když nejdříve přejdeme k praktické ukázce 
---
# TODO: možná změřit to v tomhle bodě, kolik je času na to to psát před nima, anebo ukázat a komentovat hotový kód (to je stejně asi lepší)
# ukázka z `0_flight` - nejprve `main.go` - 16:00 ??? (31:50)
Já jsem si pro vás přichystal ukázku takového jednoduchého Flightu, který se dá rovnou i vyzkoušet nebo přizpůsobit, tak se na to pojďme podívat. Budeme předpokládat, že tyto Flighty předhazujeme tomu ArgoCD pluginu, a tedy můžeme předpokládat vstupy a výstupy.

Samozřejmě jak si to naimplementujete, je vaše věc, tohle berte převážně jako příklad. První věc - inputy jsou v JSON formátu, a přijdou na vstup v `os.Stdin`, takže čteme je odtamtud. Výstupem je buď pole resourců, anebo pole *polí* resourců, v JSON formátu na `os.Stdout`. V případě pole polí se jednotlivé prvky toho vnějšího pole dají chápat jako "stage" (na pozadí to využívá ArgoCD Sync Waves anotace, čímž můžete vaše resourcy deployovat v určitém pořadí). Jakékoli chyby se pak vypisují do `os.Stderr`, což taky ukazuje tenhle jednoduchý obal hlavní funkce `run()`. Tato chyba se pak zobrazuje u chybové hlášky při failnutém Argo syncu.

V prvé řadě je dobré si definovat někde strukturu vašich vstupních dat. To bude reprezentovat jakýkoli obsah values souborů z Arga, ale taky různých override parametrů, co tam můžete nastavit. Následně tuhle strukturu dekódujete z toho příchozího JSONu. Od tohohle bodu si můžete s těmito hodnotami dělat prakticky co se vám zlíbí, ale pro případy tohoto dema to necháme jednoduché. 

Jediné co náš Flight bude vytvářet je obyčejný Deployment a Service. Necháme tady teda volání nějakých dvou funkcí `createDeployment` a `createService`, a to co nám vrátí, akorát serializujeme do JSONu a zapíšeme na `stdout`. Podívejme se na implementace těch funkcí samotných. Vracíme tady rovnou objekt `appsv1.Deployment`, a když se podívám na začátek souboru na importy, je to jen import alias nad Kube Apps API group. U těchto Kubernetích Go typů je obvyklé, že package ve kterých jsou, se jmenují podle verzí, proto je zvykem tyto aliasy nazývat takto, ale pochopitelně je to na vás. Když se podíváte na obsah toho co do tohoto Deploymentu pak vyplňujeme, tak většina se jmenuje přesně tak jak byste viděli v klasickém YAML formátu - spec, replicas, selector, template, containers, image atd. `TypeMeta` a `ObjectMeta` si zaslouží rychlý pohled na to jak jsou definované - vidíme že `TypeMeta` má u sebe json struct tag `inline` čímž JSON parserům řekne, že jeho property (`apiVerson` a `kind`) jsou na vstupu na stejné úrovni jako ostatní, ale tady je chce shluknout pod nadřazenou property `TypeMeta`. `ObjectMeta` na druhou stranu už má pouze jiný název a jedná se o klasické kubernetí `metadata`. Upřímně proč to rovnou nenazvali `metadata`, nevím.

Zbytek není třeba asi moc popisovat, a i když je to ukecané tím že musíme specifikovat pokaždé ten typ dané property, tak není s tím problém, protože IDE vám to samo napovídá co tam máte vyplnit a píše se to tabulátorem pomalu samo. Nechám si tady teda jednoduše vytvořit obyčejný Deployment v daném počtu replik s nějakým určitým labelem jako selector, pohoda. `Service` je vytvořená úplně stejným způsobem - pouze import už není z `apps` API group, ale `core`, a jinak je to stejně jako u `Deploymentu` - vyplníme API version, Kind (mimochodem doteď mě irituje že tyhle názvy nejsou nikde exportované jako konstanty, nemám tušení proč), metadata a nějakou základní specku - nezapomeneme přitom na to, aby `selector` u Service souhlasil s tím v Deploymentu. Nanejvýš si lze všimnout tady property `TargetPort` - jedná se o field, který může už ze specifikace být číslo (jako samotné číslo portu v kontejneru), anebo string, jakožto jeho název. Go samozřejmě nemá žádné type uniony jako jsou třeba v TypeScriptu, takže oni to řeší vlastním typem `IntOrString`, a akorát k němu mají vlastní custom logiku pro vytváření, čtení, serializaci a deserializaci do JSONu. 

Takže když si to tak vezmeme v rychlosti znovu jako rekapitulaci - na vstupu přečteme `stdin`, deserializujeme příchozí JSON do předem připravené známé datové struktury, uděláme s tím cokoliv je třeba, a ve výsledku posbíráme už hotové Kube resourcy a jen je zapíšeme jako pole na `stdout`. Pohoda, ne? Teďka zajímavější otázka - jak to vyzkoušet že všechno funguje?

# ukázka toho jak to otestovat 

Přece jenom je to pořád Go program, takže v klidu můžeme udělat `cat input.yaml | go run .`, a aby to bylo trochu čitelnější, proženem to ještě přes jq - `cat input.yaml | go run . | jq`. Super, všechno vypadá jak má! Někoho by mohlo napadnout, že jsem zmínil debugging, tak jak to tady udělám? Ve vscode si můžu vytvořit launch konfiguraci, kde v debuggeru můžu tenhle program pustit. Pravda, VSCode neumí předhodit uměle vytvořený `stdin`, takže pokud čekáte, že debugovat budete více, doporučuju to mírně přepsat, abyste mohli přijmout CLI flag (třeba `-f`) a v tom případě místo z `stdin` číst z toho souboru, protože CLI flagy už VSCode poslat do debuggeru umí. 

Ok, máme teda jistotu že alespoň program jako takový funguje, takže pojďme to nějak zabalit a konečně použít. V prvé řadě je třeba Flight zkompilovat do WASM. V případě Go se to dělá přes `GOOS` a `GOARCH` env proměnné - `GOOS=wasip1 GOARCH=wasm go build -o flight.wasm`. Pokud bychom si chtěli být jistí že to stále funguje jak má, můžeme použít jeden z subcommandů Yoke CLI - `yoke takeoff` (což je mimochodem ekvivalent `helm install` nebo `upgrade`) - `yoke takeoff -dry -out . test ./flight.wasm < input.yaml`

Jelikož Yoke CLI je dělané převážně na to moci tyhle Flighty přímo deployovat do Kube clusteru (jako Helm), tak je třeba dát pozor a nezapomenout `-dry` flag. `-out .` flag potom uloží všechny resourcy do YAMLů tak, jak by je deploynul do clusteru ať je můžeme zkontrolovat (anebo můžeme použít `-stdout`). Další positional argument je *název releasu* stejně jako v Helmu, a poslední je cesta k Flightu ve WASM. Tato cesta pochopitelně opět může být i remote v nějakém repozitáři, anebo jako OCI artefakt, je to fuk.

# ukázka publishe
Když máme teda jasno, nezbývá než to někam nahrát, abychom se k tomu mohli z Arga a potažmo YokeCD dostat. Yoke v tomto podporuje jednak jakýkoliv HTTP/S file hosting, a to v normálním nebo gzipped formátu, ale taky OCI registry. Já jsem zvolil Github Container Registry, což je mimo jiné i OCI registr, takže tam klidně můžeme strčit i tenhle Flight. K tomu nám slouží command `yoke stow`, který rovnou náš WASM flight gzipne a zabalí do OCI formátu (díky čemuž taky dojdeme ke slušné kompresi, z cca 16 MB na asi 3,7 MB) a pak ho tam uploadne.

# Application
Nyní už zbývá jenom adekvátně napsat náš `Application` ArgoCD resource, který pak vypadá přibližně takto. Není to už tak elegantní jako Helm, ale ta pointa je přibližně stejná - dáme tomu cestu k git repozitáři, specifikujeme pak cestu k WASM Flightu (opět, remote, OCI, ale pro fajnšmekry je tady možnost, že to Yoke dokonce zkompiluje na místě, kdybyste měli situaci že tyhle deployment věci máte hned vedle aplikačního kódu), a zbývá poslední věc - vstupy. Snažili jsme se docílit stejných možností co dokáže Helm, takže aktuálně umíme číst soubory (které se mezi sebou přepisují a prioritu má poslední v tom poli, jakožto třeba nějaký základ a overrides), ale umíme taky dávat libovolné overridy přímo tady v Application YAMLu, které dostanou ještě větší prioritu před soubory a dají se libovolně kombinovat. Bohužel, Argo CMP spec diktuje, že hodnoty v tomto objektu mohou být pouze stringy a ne libovolné datové struktury, takže musíme místy dělat trochu harakiri abychom to obešli - naštěstí YAML má ale multiline stringy, takže to není až takový problém.

No, a to je vlastně všechno, když tohle dostanete do clusteru, tak Argo Repo server pozná že se jedná o `yokecd` plugin, request na sync tímto trochu deleguje na ten plugin kontejner, ten sežvýká ty vstupní soubory a overridy, přepošle to dál na ten server kontejner, ten nám stáhne (anebo použije zacachovaný) WASM Flight, spustí ho s danými vstupy, a výstupy z toho (anebo chyby z `stderr`) přepošle zpátky až do Arga, a dále se to chová stejně jako Helm. 

---
# TODO: timing, zkusit si změřit kolik času mám na ATC, jak moc detailně to probírat
# TODO: katastrofálně - už tady jse někde na asi 31 minutách (přičemž sem se chtěl vlézt do 35)


# slide 13
Sliboval jsem, že Yoke má ještě jednu zajímavou feature, kterou Helm už nemá. Yoke se zamyslel trochu víc nad tím, že vlastně ArgoCD i Helm mají mimo jiné jeden nedostatek, a to je nedostatečně využitá validace vstupních hodnot. Myšleno, že pokud chceme deploynout nějakou aplikaci, tak nic vlastně nekontroluje validitu vstupních parametrů, a spoléháme se vlastně na implementaci takového Helm chartu (a to buď přes JSON schéma, které je stále nepovinné i když doporučované), anebo přímo nějakou validační logiku v Helm templatech. Rovněž se skrývá jakási aplikační "identita", protože Helm charty jsou prakticky nějaký obecný předpis. Identita se pak vyjadřuje metadaty jako název, anotace nebo labely. 

Yoke se rozhodl řešit právě tyto problémy a to tak, že vám dovolí vcelku dynamicky vytvářet vlastní Kubernetes resourcy, které pak implementujete vašimi Flighty jako jsme už popisovali. Pokračujíc v aviatické terminologii, jeho řešení se nazývá Air Traffic Controller, dále už jen ATC. ATC se skládá ze 2 částí - samotný ATC deployment, a tzv. Airway custom resource, která je prakticky "lepidlo" mezi vaší custom resource a její *implementací*. ATC controller je napsaný tak, že sleduje tyhle `Airway` custom resourcy, a na jejich základě jednak vytvoří odpovídající CRD (definici vaší custom resourcy), a pak v sobě spawne goroutine, která působí jako controller pro tuto resource (čili nevytváří další Pody, všechno to běží ve stejném kontejneru). Vy pak můžete přímo deploynout resource tohoto vašeho custom typu do clusteru, a ATC se postará, aby se ta resource načetla, zpracovala a výsledné resourcy se v clusteru vytvořily. Nejlépe se tahle věc vysvětluje na příkladu.

# ukázka z `1_atc_backend`
Vezměme si Flight, který jsem vám ukazoval před chvílí a máte ho ještě čerstvě v paměti. Řekněme, že bych chtěl všechny tyhle resourcy co to deploynulo (tedy Deployment a Service) shluknuté do jakéhosi funkčního sémantického celku. V našem případě to může být třeba `Backend`. V prvé řadě si tedy zadefinuju, jak takový `Backend` bude vypadat - kromě klasických metadat bych tam chtěl zachovat stejné parametry, jako jsme měli v předchozím příkladu. Vytvořím si tedy `Backend` strukturu, kde jen opíšu ty základní metadata, a pak odkaz na svůj `BackendSpec`, kde budu mít ty samotné parametry co mě zajímají.

Následně potřebuju tedy implementaci pro tuhle resource, tedy Flight, který na vstupu tentokrát nemá libovolný JSON, ale je to JSON reprezentace té `Backend` resource. Můžu tedy upravit ten Flight tak, aby místo téhle struktury deserializoval do té `Backend` struktury, a stejně tak to upravím u těch dílčích funkcí. Výstup je pak stále stejný, pole resourců jako JSON na `stdout`.

Nyní to potřebuju opět zkompilovat a nahrát někam online, pro to použiju opět ten stejný GitHub registry.

Poslední co zbývá, je vlastně vytvořit *instanci* té CRD `Airway`, díky které ATC bude konečně vědět co s tou mojí custom resourcou dělat. Můžete to udělat jako klasický YAML, ale ono je ve skutečnosti jednodušší udělat na to další malý Flight, a ukážu vám proč. Vytvořím si tedy nový soubor, který bude mít stejný obal jako ten hlavní Flight, ale jediné co na výstupu bude mít je tenhle `Airway` objekt. Podíváme se, co všechno tady můžeme specifikovat:
- jako název Airwaye si vybereme vlastně název té cílové CRD, s tím že API group je v podstatě na vás - třeba `nginxes.example.com`
- následuje `AirwaySpec`, kde prvně zvolíme *mód* - prozatím vyplníme `Standard`, později se ještě k tomuhle vrátím
  - `WasmURLs` definuje vlastně cestu k našemu implementačnímu Flightu, proto jsme to museli někam nahrát nejprve
  - `Template` pak už obsahuje přesně to, co byste vyplňovali v klasickém CRD (jde to vidět i z toho Go typu, že jde o `CustomResourceDefinitionSpec`
    - `Group` je nějaký vámi zvolený název API group
    - `Names` pak definuje název té custom resource
    - `Scope` řekneme, jestli je to namespaced resource nebo cluster resource
    - následuje pak pole CR verzí, a tady přesně vidíte důvod, proč jsem to doporučoval opět jako Flight - Yoke nám dává k dispozici pěknou utility funkci, která dokáže reflexí zkonstruovat OpenAPI schéma z Go struktur
      - nutno podotknout, že ne všechno je schopná tahle funkce "přeložit", např. typy s custom marshal logikou, nebo `oneOf` typy apod.
      - když byste měli něco složitějšího, museli byste si to asi vyplnit ručně, a to už je jednodušší to napsat možná přímo v YAMLu
    
Tento mini Flight pak opět zkompilujeme - `GOOS=wasip1 GOARCH=wasm go build -o airway.wasm`, a můžeme rovnou deploynout do clusteru - `yoke takeoff -wait 5m backend_airway ./airway.wasm`. Anebo, pokud to chcete mít taky nějak verzované v gitu, můžete si z toho vyrenderovat jen ten cílový YAML přes `-dry` flag, a ten pak commitnout a nechat nasadit třeba Argem, to už je na vás.

ATC potom za vás vytvoří `Backend` CRD a tedy spawne goroutine, která tyhle objekty umí zpracovávat, a dovolí vám deployovat aplikace tím, že aplikujete pouze tyhle `Backend` objekty. 

Osobně si myslím, že je to docela cool feature, protože sám moc dobře vím, jaká bolest je psát tyhle věci "klasicky" - ručně si definovat tyhle Custom Resourcy, a pak psát s pomocí kubebuilderu, controller-manageru a jiných jako vlastní controller, a řešit jeho nasazení, reconcile loop a spoustu dalších věcí. Všechny tyhle nepohodlnosti za vás ATC docela slušně abstrahuje, a můžete se více soustředit na tu samotnou byznys logiku za tím.
---
# slide 14
Říkal jsem, že se vrátím ještě k těm módům těch Airwayů. Yoke má aktuálně 3 módy a liší se prakticky v tom, co se s těmi deploynutými věcmi děje *po nasazení*. Základním a defaultním módem je právě `Standard`, který vlastně říká, že ty resourcy jsou po nasazení unmanaged. Což znamená, že potom co se to nasadí, tak je kdokoliv může různě editovat, a ATC to nebude nějak řešit nebo napravovat to (tedy v podstatě jako klasický Helm release, taky nasadíte a od té doby ciao, už to není můj problém). Tento stav pak přetrvává do doby, dokud někdo nezmění samotnou tu instanci té resource (ne Airway, ale třeba náš `Backend` resource), v tom případě dojde ke spuštění reconcile, ATC nechá vygenerovat nový stav těch child resourců, a aplikuje je do clusteru, čímž se jakékoli konfliktní změny vrátí do požadovaného stavu. Pokud by člověk chtěl zachovat nějakou samoopravu a udržovat více ten "zamýšlený" stav jako to dělá ArgoCD při zapnutém Auto-Syncu, pak je tady property `fixDriftInterval` ve specifikaci Airwaye.

Druhý mód je `Static`, který dělá to, že po nasazení vysloveně *blokuje* všechny úpravy svých child resourců pomocí admission webhooku. Může se tedy hodit pro případy, pokud absolutně nechcete, aby vám do těch child resourců kdokoliv hrabal.

Poslední mód je obzvláště zajímavý, a to je `Dynamic`. Zatímco ve Static a Standard módu k "přepočítání" stavu těch resourců dojde pouze v případě že se změní ta custom resource (anebo fixně na nějakém intervalu), tak u Dynamic módu k tomu dojde při jakékoliv změně jakýchkoli child resourců. Což jednak může znamenat to, že k "nápravě" stavu dojde prakticky hnedka (ale v tom případě se nijak neliší od Static módu). Na druhou stranu, v Dynamic módu nemusíte z Flightu vracet pokaždé *to samé*.

Jednu věc jsem totiž nezmínil, Yoke v kontextu ATC a naopak přímého nasazování přes Yoke CLI má totiž *konfigurovatelný přístup* k tomu clusteru (podobně jako Helm má právě `lookup()` funkci. Nebojte se, není třeba hned bít na poplach, že je to nebezpečné, je to jednak omezené na pouze GET na jedinou resource (ne list request), a jednak je granulárně konfigurovatelná ve smyslu "co přesně můžu číst". Můžu si třeba nastavit, že můj Flight má práva pouze na čtení ConfigMap z určitého namespace, anebo dokonce to omezit i na jednu jedinou ConfigMapu podle názvu. 

Kombinace právě téhle lookup funkce, toho že se v Dynamic módu spouští reconcile při updatu každé "sledované" resourcy a toho že nemusím pokaždé vracet to stejné pak znamená, že se dá dynamicky reagovat na změny nebo nějaký stav v clusteru - můžu implementovat nějakou formu orchestrace. Klasickým příkladem může být, když třeba potřebujete nějaké objekty vytvářet postupně, až po uplynutí nějaké doby nebo splnění nějaké podmínky. Například si můžete nechat vytvořit databázi někde v cloudu, vysloveně *počkat* až její controller zapíše credentials pro přístup k ní typicky do nějakého Secretu, a pak ve vašem Flightu tento Secret referencovat a mountnout jako env proměnné. Že jako nejde o případ tzv. eventuální konzistence, kdy se ten Deployment prakticky točí v nějakém vadném nefunkčním stavu *dokud* ta databáze neexistuje, a pak se sám nějak "chytne". Kdybyste tohle chtěli dělat bez Yoku, prakticky si tyhle podmínky musíte pohlídat buď ručně, nebo psát na to nějaké skripty, anebo (v nějakém složitějším případě) pro to vysloveně psát custom controller.

# TODO: opět timing, rád bych tuhle věc ukázal aspoň rychle na příkladu, ale asi nebude čas
---
# slide 15
No, tohle byl vcelku vyčerpávající výčet toho, co Yoke umí, tak pojďme si dát dohromady nějaké srovnání obou manažerů. Nejprve bych vypíchnul nějaké pro a proti jednotlivých technologií a na závěr je nějak srovnal. Začněme s Helmem:
- ruku na srdce, je battle tested - je v provozu asi asi 9-10 let, od roku 2018 je pod taktovkou CNCF a plně akceptovaný jako "Graduated" projekt byl v roce 2020. Používají ho stovky firem a tisíce uživatelů po celém světě
- ve většině případů se poměrně jednoduše používá - deployment libovolného Chartu je otázkou cca 1-3 commandů, i napsat si vlastní jednodušší chart není nic složitého
- má nativní support v ArgoCD i jiných systémech, např. FluxCD
- má možnost používat sub-charty pro nasazení složitějších aplikací
- můžete použít i tzv. Helm hooky pro implementaci složitějších nasazovacích mechanismů/akcí

Na druhou stranu:
- dá se uvažovat, že není úplně bezpečný, z hlediska že není tak jednoduché nějak zkontrolovat co přesně si nasazujete, nebo můžou být charty podvrženy
- v případě složitých chartů, co obsahují spoustu logiky, tak ta udržovatelnost jde hodně ke dnu, a absence nějakých vestavěných testovacích mechanismů to jen zdůrazňuje
---
# slide 16
Yoke oproti tomu:
- má vcelku slušnou zpětnou kompatibilitu s Helmem umožňující plynulejší přechod
- využívá klasických programovacích jazyků ke tvorbě Kube resourců, namísto textových YAML template
- díky tomu má plnou podporu IDE, typovou bezpečnost, testování atd.
- má možnost aplikace reprezentovat díky ATC ve formě custom resourců
- rovněž má možnost implementovat složitější mechanismy než Helm právě díky zmíněné orchestraci v Dynamic módu

Na druhou stranu:
- částečně trpí podobnými problémy jako Helm - pořád si můžete deploynout něco jiného než si myslíte
  - na druhou stranu, veřejně dostupné Flighty jsou prakticky neexistující, což se rozhodně nedá říct o Helm chartech
- jako osobně největší potenciální mínus vidím to, že je to velice nový projekt (první commit v lednu 2024), který nemá žádné sponzory, a prakticky jednoho maintainera
  - pro některé toto může být důvod to jako možnost zamítnout, ale osobně si myslím, že i v tom stavu v jakém to je teď, je to vice než schopné a stačilo by udělat třeba fork nebo kopie docker imagí apod., pokud byste měli strach o to, že by to autor zabalil
---
# slide 17
A nakonec tedy trochu subjektivní srovnání obou:
- oba dva umí spravovat svoje aplikace v clusteru napřímo přes CLI
- oba mají ve výsledku podporu v ArgoCD a jsou schopny dosáhnout stejných výsledků 
- oba mají stejné bezpečnostní problémy ve formě neověřených resources, imagí, závislostí nebo default hodnot, ale v tomhletom ohledu to Helm schytá stoprocentně hůř, protože má rozsáhlý katalog takto neověřených Chartů. Yoke tímhle trpí spíše teoreticky

- Helm mi osobně přijde jednodušší v případě nějakých jednodušších chartů - když tam není moc logiky, YAML templaty se stále jednodušeji čtou a píšou
- Helm je rozšířenější a má větší podporu v jiných nástrojích (např. FluxCD)

- Yoke má nesrovnatelně lepší možnosti pro komplexní charty právě v tom, že používá klasický programovací jazyk a má plnou podporu IDE
- Yoke má zajímavý celý ten koncept ATC, že můžete aplikace reprezentovat custom resourcy (na které mimochodem taky jdou aplikovat RBAC práva v clusteru), a dokonce nad nimi psát i složitou dynamickou logiku (a tím abstrahuje poměrně složitý proces implementace vlastního controlleru)
- osobně považuju za plus (tentokrát ne v kontextu srovnávání) i to, že Yoke vás netlačí do toho používat ho jedním nebo druhým způsobem. Dává vám hned několik možností, jakým ho můžete využívat v závislosti na vašich vlastních use casech. Například pro nás v ProRocketeers, nám bohatě vyhovuje mít Yoke jako ArgoCD plugin, a jsme s tím spokojení. Ale jsem rád, že vím, že kdybych potřeboval něco složitějšího, mám tady tu možnost využít ATC aniž bych se musel patlat s vlastním controllerem.
---
Tohle je z mé strany asi vše, děkuji vám všem za pozornost a nechávám vám prostor na případné otázky
...
Pokud byste měli další otázky, které už bychom nestihli probrat, najdete mě ještě chvilku tady v prostorách před místností, v opačném případě neváhejte se mi ozvat na mail `lukas.stuchlik@prorocketeers.com`.