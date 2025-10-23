# slide 1 - 0:08
Dobrý den, jmenuju se Lukáš Stuchlík a rád bych vám tady něco řekl o alternativě k Helmu, aneb jak se nezbláznit z YAML templatů

# slide 2 - 0:40 (0:48)
V kostce o mně, vystudoval jsem magisterské studium na Ostravské univerzitě v oboru Informační systémy, a od té doby dělám už 5 let pro ostravskou firmu ProRocketeers. Jako firma děláme agilní vývoj softwaru pro klienty, kde převážně nabízíme Team as a Service. Já osobně se věnuju částečně nějakému backend vývoji v Typescriptu a Go, ale převážně jsem jako DevOps Engineer - správa CI/CD pipelines, Kubernetes clusterů, dělám s Helmem, Dockerem, Ansible apod. Ve volném čase se rád věnuju tomu přispět něco do opensource komunity zpátky, mimo jiné i do projektu, který vám tady budu představovat

# slide 3 - 1:13 (2:02)
Téma mojí přednášky je předvést alternativu pro Helm, tak se pojďme nejprve podívat na něj. 

Helm je package manažer pro Kubernetes. Jeho ústředním konstruktem jsou tzv. Charty, což je obvykle pouze kolekce YAML šablon, které krmíme parametry (Helm values), někdy i s přidanými závislostvmi ve formě jiných sub-chartů. Když pak zavoláme Helm, aby nám takový Chart nasadil, tak posbírá values z různých zdrojů - obvykle to bývají nějaké default hodnoty přímo uvnitř Chartu, a pak různé vaše overridy. Typicky se takhle předávají celé soubory (tzv. values fily), ale Helm si rád poradí i s CLI argumenty, kam se dají protlačit třeba env proměnné apod.

Tyto parametry pak spolu s YAML šablonami jdou přes Helm templating engine, který je založený na jazyce Go, a prakticky takhle dosadí ty vaše parametry na příslušné místa v těch šablonách a výsledkem z toho jsou Kube objekty v YAML formátu. Helm vám taky umožňuje do jisté míry nějaké logické věci, jako třeba flow control, proměnné, manipulace s poli, objekty, stringy atd. Má to však i své stinné stránky...

# slide 4 - 1:32 (3:34)
U nás v ProRocketeers míváme i různé malé služby a projekty, které jsou pod naší kontrolou a potřebují taky nasadit na některé z našich vlastních clusterů. Dříve jsme to řešili typicky tak, že jsme používali předgenerované Helm charty co vám vyhodí příkaz "helm create", které jsme si upravovali dle potřeb daného projektu. Tento přístup ale byl strašně repetitivní, a když jsme se rozhodli dělat nějaké změny globálně napříč všemi aplikacemi, dalo to vždycky hroznou spoustu práce to upravit všude. Vzal jsem si tedy na sebe nelehký úkol, a dal se do přípravy vlastního "obecného" Helm chartu, který bude centralizovaný a společný pro všechny takovéhle projekty, a který obsáhne všechno, co tak může běžná aplikace potřebovat. A že jich ve výsledku není úplně málo, mezi běžně používané věci můžu vyjmenovat třeba velkou konfiguraci Deploymentu samotného, ale i věci okolo, např. Ingress, ExternalSecrety, HorizontalPodAutoscaler, RBAC práva pro ServiceAccounty a plno jiných věcí.

Vzhledem k tomu, že jsem k tomu chtěl mít jaksi.. hezký formát těch values, aby i developer experience (tedy naše) s používáním tohoto chartu byla přívětivá, nevyhnutelně přišla na řadu dost velké množství logiky. Různý preprocessing těchto values, validace, upravení těch vstupních hodnot do formátu, ve kterém se zase tomu template enginu s tím lépe pracuje atd. A přesně tady začínají pramenit problémy se správou Helm chartů.

# slide 5 - 0:12 (3:46)
Na obrázku můžete vidět zhruba do jakého stavu jsem se dostal. Pořád se to dá vcelku číst, ale vyvíjet to je daleko horší. 

# slide 6 - 2:01 (5:48)
Podpora IDE pro Helm templaty je dost mizerná, sotva se mi podařilo najít nějaký extension pro syntax highlighting, ale na takové základní věci jako IntelliSense pro napovídání datové struktury těch objektů jsem už mohl zapomenout a místo toho to držet v hlavě nebo někde bokem. S tímto ještě půl bídy, chybové hlášky v případě že se snažíte dereferencovat proměnnou, která třeba neexistuje jsou ještě vcelku ok, tam vám Helm alespoň řekne nějaké vodítka, třeba název toho co se snažíte přečíst..

Horší problém pramení z toho, že těmi šablonami vytváříte YAML soubory, a ty jsou závislé na odsazení - whitespace. Helm templaty, resp. ty dynamické části, které píšete do dvou složených závorek, totiž mají na každé straně optional pomlčku, a ta tomu enginu říká, že z téhle strany "spolkni" veškerý whitespace až do prvního normálního znaku. A řeknu vám, když koukáte do stovek řádků takovýchto template, tak je *strašně jednoduché* někde takovou pomlčku buď zapomenout, anebo ji dát někam, kde být nemá. A nikdo vás na to neupozorní, že je něco špatně až na to, že vám začne render padat na YAML syntax chybách, ze kterých nejde poznat, kde je co špatně. Nezbývá vám pak nezbývá nic moc jiného než pozorně pročítat řádek po řádku a hledat, kde je zdroj takové chyby. 

Další problém - řekněme, že si chcete ušetřit budoucí nervy, a rozhodnete se zavést nějaké testování, ideálně unit testování do svého Chartu. Ne že by to nešlo, existuje na to Helm plugin - `helm-unittest`, kdy i ty samotné unit testy píšete v YAMLech, ale ani práce s ním není úplně bezchybná (např. jsem se dostával do nekonzistencí, že lokálně při vývoji mi testy procházely, ale v CI/CD pipeline už ne a nepřišel jsem na to proč). 

Dlouho mi tyto problémy ležely v žaludku a chtěl jsem něco, čím bych to nahradil, a to mě přivádí k naší alternativě, kterou vám chci představit - Yoke.

# slide 7 - 1:20 (7:08)
Yoke se představuje taky jako package manažer, ale *jinak*. Místo textových YAML šablon se rozhodl jít jinou cestou, a tyhle Kube objekty (často nazývané resourcy) vytváří v klasickém programovacím jazyce, čímž se o něco víc blíží pojmu "infrastructure as code". Pojďme si to trochu přiblížit. Jelikož je to pointou celé mojí prezentace, rovnou to budu dávat do srovnání s Helmem, který je lidem většinou známý.

Začněme u Chartů - jejich ekvivalent v Yoku se jmenuje Flight. Yoke celkově bere svoje názvosloví z aviatiky - netuším moc proč, možná má autor prostě rád letadla? Kdoví. Flight je obecně řečeno program v libovolném jazyce, který je zkompilovaný do WebAssembly formátu, a který na vstupu bere libovolné parametry, a na výstupu dává Kubernetes resourcy. Proč zrovna WebAssembly? Z několika důvodů: 

1) Je to společný kompilační cíl pro více různých jazyků
2) Je nezávislý od cílového OS nebo architektury, protože produkuje neutrální bytekód podobně jako Java
3) Jeho runtime je sandboxed, tj. nemá přístup k souborovému systému anebo síti

# slide 8 - 2:02 (9:11)
I když můžete teoreticky psát Flighty v různých jazycích, je tak nějak obecně doporučované vybrat si právě Go, a to z pár důvodů - jednak Yoke distribuuje sadu pomocných SDK právě pro Go, ale hlavní důvod za mě je oficiální podpora Kubernetes API typů, které si importujete do vašeho Flightu jako balíčky. Tím rovnou narážím na (z mého pohledu) hlavní výhodu, a to je plná podpora programovacího jazyka vaší volby. Můžete si svůj Flight napsat jako klasický Go program, používat všechny featury tohoto jazyka (i to co vám Helm v templatách nedovolil), máte k tomu plnou podporu IDE, klasickou type safety (a to nejen ve vašich datových strukturách ale právě i v Kube API strukturách), unit testing, dokonce i klasický breakpoint debugging. 

Dalo to nějakou práci ten Chart přepsat do Yoke implementace, ale ta jistota, kterou jsem tímto refactorem nabyl, se mi zatím zdá dost k nezaplacení. 

Nicméně možnosti Yoku nekončí tady. Dejme tomu, že ve vašem Helm Chartu máte i nějaké sub-charty jako závislosti (např. Postgres databázi nebo Redis), a do jejich implementace už šahat nemůžete. Yoke na tyhle situace reaguje tak, že má vcelku slušnou kompatibilitu s Helmem. 

1) Můžete zabalit existující Helm chart, a renderovat ho přímo z Yoke Flightu (pomocí Go direktivy `go:embed`, přičemž se z toho archivu stane pole bajtů v paměti)
2) Anebo můžete vyzkoušet automaticky *vygenerovat* Yoke Flight z Helm Chartu jako takového - `helm2go`
  - tenhle CLI tool opět zabalí existující Helm chart, ale navíc k tomu zkusí z JSON schématu Chartu vygenerovat Go struktury, které to reprezentují, čehož můžete taky využít dle libosti
  
# slide 9 - 0:17 (9:28)
Tady v rychlosti, jak pak úryvek takového programu může vypadat..

# slide 10 - 1:17 (10:45)
Zatím jsme se tady bavili hlavně o tom, co Yoke dokáže, ale pojďme se podívat teďka na to, jakým způsobem oba tooly (Helm i Yoke) fungují ve světě GitOps, specificky právě ArgoCD.

Helm je vysloveně nativně integrovaný do Arga, bez jakékoliv konfigurace. Můžete tam specifikovat vstupní parametry, číst je ze souborů, ale taky inline v Application manifestech. Charty můžete používat lokálně z připojeného Git repozitáře, nebo z Helm repozitáře, nebo jako OCI artefakt. Argo Helm využívá ovšem pouze na ten *templating*, což ho v pár věcech trochu omezuje.

Oproti tomu Yoke je v tomhle světě něco trochu cizího. Naštěstí ArgoCD má koncept CMP - Content Management Plugin. Je to způsob, jakým si můžete integrovat libovolný způsob, kterým renderujete Kube resourcy a o zbytek se už Argo postará jako dříve. Z hlediska DX je to pro nás trochu nepohodlnější, ale bohužel s tím už nic neuděláme, to má pod taktovkou ArgoCD a ne Yoke. Dobrá zpráva je, že i s omezeními, které nám diktuje CMP specifikace, jsme stále schopni dosáhnout prakticky stejného výsledku jako Helm.

# slide 11 - 1:51 (12:36)
Technická realizace Yoke pluginu pro ArgoCD (nazývaného YokeCD) je takováto:
- ArgoCD samotné se deployuje do clusteru jako asi 6 komponent, z čehož jedna je "repo-server"
- CMP specka diktuje, že pluginy se přidávají do Arga jako sidecar kontejnery vedle hlavního "repo-server" kontejneru, přičemž použijou společnou binárku pro start CMP serveru + Plugin specifikaci, kterou Argo umí přečíst
- YokeCD je ovšem ve své implementaci trochu složitější, a tak se do Podu přidává další kontejner ("yokecd-svr"), což je jednoduchý HTTP server, který komunikuje s YokeCD plugin kontejnerem
    - tato spolupráce spočívá v tom, že yokecd plugin kontejner působí jako prostředník, jakási fasáda mezi repo serverem (se kterým komunikuje přes gRPC), a tím yokecd serverem (přes HTTP)
    - yokecd plugin pak od repo serveru přijímá input parametry, které různě zpracuje a pošle dál na server kontejner
    - ten se postará o stažení a caching našeho Flightu, jeho samotné spuštění a tím render těch Kube resource, které pak zpětně vrátí až do repo serveru

Je to trochu komplikovanější než by bylo libo, ale výhodou je to, že i navzdory tomu, že Flight je i v zabaleném formátu třeba 15MB soubor, tak je render díky cachování stále velmi rychlý.

Yoke má oproti Helmu navíc ještě jednu zajímavou feature, ale pro její představení bude lepší, když nejdříve přejdeme k praktické ukázce 

# slide 12 - 2:51 (15:27)
Já jsem si pro vás přichystal ukázku takového jednoduchého Flightu, ať máte reálnou představu, jak se to může implementovat.

První věc - inputy jsou v JSON formátu, a přijdou na vstup v `os.Stdin`, takže čteme je odtamtud. Výstupem je buď pole resourců, anebo pole *polí* resourců, v JSON formátu na `os.Stdout`. V případě pole polí se jednotlivé prvky toho vnějšího pole dají chápat jako "deployment stage" (na pozadí to využívá ArgoCD Sync Waves anotace, čímž můžete vaše resourcy deployovat v určitém pořadí). Jakékoli chyby se pak vypisují do `os.Stderr`, což taky ukazuje tenhle jednoduchý obal hlavní funkce `run()`. Tato chyba se pak zobrazuje u chybové hlášky při failnutém Argo syncu.

Nejprve je dobré si definovat někde strukturu vašich vstupních dat. To bude reprezentovat jakýkoli obsah values souborů z Arga, ale taky různých override parametrů, co tam můžete nastavit. Následně tuhle strukturu dekódujete z toho příchozího JSONu. Od tohohle bodu si můžete s těmito hodnotami dělat prakticky co se vám zlíbí, ale pro případy tohoto dema to necháme jednoduché. 

Jediné co náš Flight bude vytvářet je obyčejný Deployment a Service. Já ponechal jako ukázku pouze ten Deployment, protože Service se pak dělá prakticky ekvivalentně a není třeba to tu ukazovat. Máme teda nějakou funkci, která nám bude na vstupu brát ty naše parametry, a vracet nám bude Kubernetes objekt Deployment. Tyhle typy rovnou importujeme z Kubernetes API, a vyplníme je pouze nějakými základními údaji, např. tady název, namespace do kterého patří.. Dále už pak vyplňujeme specku toho Deploymentu, takže ostatní věci jako počet replik, selector na Pody (který vzhledem k tomu, že se musí shodovat s tím, co budeme mít v Service, rovnou můžu vytáhnout do pomocné funkce `commonLabels`), a nějaký jednoduchý kontejner. Jak jsem už říkal, výhodou tady je to, že vám IDE v každém bodě napovídá, co tam má být a nemusíte si tyhle typy nebo property pamatovat. 

No, a to je vlastně všechno z hlediska implementace. Máme maličký program, který bere něco na vstupu, sežvýká to a vrátí nám Kubernetes objekty na výstupu. Teďka zajímavější otázka - jak to vyzkoušet že všechno funguje?

# slide 13 - 1:21 (16:49)
Pořád je to klasický program, takže to můžeme normálně spustit s `go run .`

Pro potřeby debuggingu doporučuju přiimplementovat možnost předat tomu vstupní soubor, lépe s tím umí pracovat třeba VSCode

Nebo to můžeme zkompilovat s adekvátními parametry, a pak využít subcommandu Yoke CLI `yoke takeoff`

Nakonec je třeba to někde publishnout, s čímž nám pomůže `yoke stow`, který ten WASM Flight zabalí jako OCI artefakt a uploadne do nějakého úložiště.

# slide 14 - 1:02 (17:51)
Nyní už zbývá jenom adekvátně napsat náš `Application` ArgoCD resource, který pak vypadá přibližně takto.. Není to už tak elegantní jako Helm, ale ta pointa je přibližně stejná - dáme tomu cestu k git repozitáři, cestu k tomu Flightu, a pak specifikovat vstupy. 

Snažili jsme se docílit stejných možností co dokáže Helm, takže aktuálně umíme číst soubory, ale umíme taky dávat libovolné overridy přímo tady v YAMLu, které dostanou ještě větší prioritu před soubory a dají se libovolně kombinovat. 

No, a to je vlastně všechno, když tohle dostanete do clusteru, tak Argo Repo server pozná že se jedná o `yokecd` plugin, request na sync tímto deleguje na ten plugin kontejner, ten sežvýká ty vstupní soubory a overridy, přepošle to dál na ten server kontejner, ten nám stáhne (anebo použije zacachovaný) Flight, spustí ho s danými vstupy, a výstupy z toho (anebo chyby z `stderr`) přepošle zpátky až do Arga, a dále se to chová stejně jako Helm. 

# slide 15 - 1:47 (19:38)
Sliboval jsem, že Yoke má ještě jednu zajímavou feature, kterou Helm už nemá. Yoke se zamyslel trochu víc nad tím, že vlastně ArgoCD i Helm mají mimo jiné jeden nedostatek, a to je nedostatečně využitá validace vstupních hodnot. Myšleno, že pokud chceme deploynout nějakou aplikaci, tak nic vlastně nekontroluje validitu vstupních parametrů, a spoléháme se vlastně na implementaci takového Helm chartu (a to buď přes JSON schéma, které je stále nepovinné i když doporučované), anebo přímo nějakou validační logiku v Helm templatech. 

Yoke se rozhodl řešit právě tyto problémy a to tak, že vám dovolí vcelku dynamicky vytvářet vlastní Kubernetes resourcy, které pak implementujete vašimi Flighty. Pokračujíc v aviatické terminologii, jeho řešení se nazývá Air Traffic Controller, dále už jen ATC. ATC se skládá ze 2 částí - samotný ATC deployment, a tzv. Airway custom resource, která je prakticky "lepidlo" mezi vaší custom resource a její *implementací*. ATC controller pak sleduje tyhle `Airway` objekty, a na jejich základě jednak vytvoří odpovídající CRD (definici té vaší custom resourcy), a pak v sobě spawne goroutine, která působí jako controller pro tuto resource. Vy pak můžete přímo deploynout resource tohoto vašeho custom typu do clusteru (třeba klidně i přes Argo), a ATC se postará, aby se ta resource načetla, zpracovala a výsledné resourcy se v clusteru vytvořily.

# slide 16 - 1:33 (21:12)
Implementačně na příkladu to není až tak rozdílné oproti klasickému Flightu, prakticky jediné co se změní je to, že trochu více formalizujete specifikaci vstupů (přidají se povinné Kubernetes metadata a status) a samozřejmě přepíšete váš kód aby načetl a pracoval s těmito strukturami.. A hlavně sepíšete *instanci* té `Airway` CRD. Tady je další dobrý důvod proč používat Go, protože Yoke vám nabídne možnost reflexí vygenerovat OpenAPI schéma z vašich Go struktur. Tahle Airway se pak může opět zkompilovat a deploynout jako maličký Flight (který ani nemusí být nikde nahraný), ale můžete si z toho nechat vygenerovat YAML který nasadíte jiným způsobem, jak chcete.

Osobně si myslím, že je to docela cool feature, protože sám moc dobře vím, jaká bolest je psát tyhle věci "klasicky" - ručně si definovat tyhle Custom Resourcy, a pak psát s pomocí kubebuilderu, controller-manageru a jiných jako vlastní controller, a řešit jeho nasazení, reconcile loop a spoustu dalších věcí. Všechny tyhle nepohodlnosti za vás ATC docela slušně abstrahuje, a můžete se více soustředit na tu samotnou byznys logiku za tím.

# slide 17 - 1:06 (22:18)
Yoke ATC má navíc ještě pár tzv. módů, kterým ovlivňuje, jak se Yoke chová k jeho child resourcům potom co se nasadí. Nebudu to tady rozebírat moc detailně, takže jen v kostce:
- `Standard` - je to prakticky to co dělá Helm - po nasazení se ty resourcy nijak nehlídají a můžete je i upravit dle libosti, mění se až po změně té CR
- `Static` - po nasazení se objekty zamrazí a jakékoliv změny jsou admission webhookem blokované
- `Dynamic` - sledují se deploynuté ale i libovolné jiné resourcy a Flight se může svou funkcionalitou přizpůsobit *aktuálnímu* stavu clusteru - dovoluje to implementovat komplexní logiku i nějakou orchestraci, jako byste čekali v klasickém controlleru

# slide 18 - 1:23 (23:41)
No, tohle byl vcelku vyčerpávající výčet toho, co Yoke umí, tak pojďme si dát dohromady nějaké srovnání obou technologií. Začněme s Helmem:
- ruku na srdce, je battle tested - je v provozu asi asi 9-10 let, od roku 2018 je pod taktovkou CNCF a plně akceptovaný jako "Graduated" projekt byl v roce 2020. Používají ho stovky firem a tisíce uživatelů po celém světě
- ve většině případů se poměrně jednoduše používá
- má nativní support v ArgoCD i jiných systémech, např. FluxCD
- má možnost používat sub-charty pro nasazení složitějších aplikací
- můžete použít i tzv. Helm hooky pro implementaci složitějších nasazovacích mechanismů/akcí

Na druhou stranu:
- dá se uvažovat, že není úplně bezpečný, z hlediska že není tak jednoduché nějak zkontrolovat co přesně si nasazujete, nebo můžou být charty podvrženy
- v případě složitých chartů, co obsahují spoustu logiky, tak ta udržovatelnost jde hodně ke dnu, a absence nějakých vestavěných testovacích mechanismů to jen zdůrazňuje

# slide 19 - 1:26 (25:08)
Yoke oproti tomu:
- má vcelku slušnou zpětnou kompatibilitu s Helmem umožňující plynulejší přechod
- využívá klasických programovacích jazyků ke tvorbě Kube resourců, namísto textových YAML template
- díky tomu má plnou podporu IDE, typovou bezpečnost, testování atd.
- má možnost aplikace reprezentovat díky ATC ve formě custom resourců

Na druhou stranu:
- částečně trpí podobnými problémy jako Helm - pořád si můžete deploynout něco jiného než si myslíte
  - na druhou stranu, veřejně dostupné Flighty jsou prakticky neexistující, což se rozhodně nedá říct o Helm chartech
- jako osobně největší potenciální mínus vidím to, že je to velice nový projekt (první commit v lednu 2024), který nemá žádné sponzory, a prakticky jednoho maintainera
  - pro některé toto může být důvod to jako možnost zamítnout, ale osobně si myslím, že i v tom stavu v jakém to je teď, je to vice než schopné a stačilo by udělat třeba fork nebo kopie docker imagí apod., pokud byste měli strach o to, že by to autor zabalil

# slide 20 - 0:10 (25:18)
Tohle je z mé strany asi vše, děkuji vám všem za pozornost a nechávám vám prostor na případné otázky
...
Pokud byste měli další otázky, které už jsme nestihli probrat, najdete mě ještě chvilku tady v prostorách před místností, v opačném případě neváhejte se mi ozvat na mail `lukas.stuchlik@prorocketeers.com`.