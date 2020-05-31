
# Audit System Chain Code (ASCC) : An Add-On Plugin for Peer


# Installation
1. ASCC 빌드하여 ascc.so 생성 (build.sh)
2. ascc.so 를 지정된 위치로 복사 (deploy.sh)
3. (처음인 경우) 피어 도커 이미지 실행 시의 옵션 설정: .yaml 에서 volumes 맵핑 추가하기. 맵핑 시에 2에서 지정된 위치를 /opt/lib/ 로 설정해야 함
<!-- 4. (처음인 경우)  -->

# Test
1. 기존에 띄어진 컨테이너 있으면 재시작시키면 된다. (docker restart peer0.org1.example.com)

# Note
- 테스트 시에, Init 에 Put 하는 경우, "no ledger context" 에러 메시지 발생.

# Data Store
- DS["txid"] = AuditTx // store AuditTx itself
- DS["block.number.member.id"] = txid // block based filtering
- DS["member.id.block.number"] = txid // member based filtering

```
environments:
  - FABRIC_LOGGING_SPEC=ascc=debug

volumes:
  - {ASCC_PATH}/ascc/:/opt/lib/
```
