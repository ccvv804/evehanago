# evehana (golang)
hana ICM audio extractor (golang)
<pre><code>go build eve.go
eve.exe -file sample.ICM
</code></pre>
hana ICM 오디오를 wav로 변환해주는 추출기 입니다. 

ICM 파일은 HANA/SKY 시리즈 기기의 오디오 파일이나 KY? 파일에서 추출하여 구할 수 있습니다. 

KY? 파일에서 추출하고 싶다면 [Dummy K-Chorus Ripper](https://github.com/ccvv804/dkcr)를 사용할 수 있습니다.

sample.ICM은 Mike Koenig가 [soundbible](https://soundbible.com/1003-Ta-Da.html)으로 업로드한 wav파일을 ICM으로 변환했습니다. sample.ICM 파일 인코딩에는 [wav2icm](https://github.com/ccvv804/wav2icm)가 사용되었습니다.
