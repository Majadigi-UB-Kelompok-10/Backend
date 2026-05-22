#!/bin/bash

# USES RELATIVE PATH, SHOULD BE RUN IN CURRENT FOLDER
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Local Variable for Substitution
# -- Services --
export BANSOS_UUID="0136797f-e16b-4f8a-a9a2-1fd90e135ea3"
export BAPENDA_UUID="94b02f22-9c4d-461a-883b-046b78975d5c"
export JDIH_UUID="324f2278-f7b8-4680-bfdc-17015673006a"
export KLINIK_UUID="4bbab6bb-3ea4-4d51-a23e-08382ce5c32b"
export RSSA_UUID="2f5f0476-c1f1-4bb8-9ce9-0adc2cf22da2"
export SIDITA_UUID="48b34e1e-2f4c-48c3-9ae0-96cdc5326311"
export SINAKER_UUID="a9e7bcc7-7b6d-47e7-9d04-daa0f030110c"
export SISKAPERBAPO_UUID="656ed163-4303-4c94-895b-cfa0d634c703"
export TRANSJATIM_UUID="d9f92735-903e-4f5d-97c0-627755ded150"

# -- Endpoints --
# -- Bansos --
export SAPABANSOS_UUID="b6ac909c-f6e8-4cc6-b0cc-c032ae545daf"

# -- Bapenda --
export PKB_UUID="776d2206-2b78-4857-9494-523b6de049cf"
export NJKB_UUID="a03b781a-b8a2-4a5b-847e-b777a0ff5d35"

# -- JDIH --
export PHJT_UUID="f3f14045-731b-4f03-8ac7-8b67f6ce69e0"

# -- Klinik Hoaks --
export KH_UUID="418f63c4-c769-4f22-bbd3-2a095ac2de8f"

# -- RSSA (Rumah Sakit Dr. Saiful Anwar) --
export KKR_UUID="a7310ace-aa69-4259-a299-465e096e1295"

# -- SIDITA (Sistem Informasi Daya Tarik Wisata Jawa Timur) --
export SIDI_UUID="6d4c3cca-11df-4864-981d-84807bbfdb95"

# -- SINAKER (Sistem Informasi Ketenagakerjaan) --
export DPK_UUID="03a25ee2-2a13-4bb3-a122-c83a5ebcb297"
export BLK_UUID="ded52de0-c3ec-45d3-9bbf-d70b80b69b4c"
export CPP_UUID="4cb1d1ea-e88e-46f7-b0bf-97dd2a9fe693"

# -- SISKAPERBAPO --
export HBP_UUID="9d4a9d2b-55b4-411d-b1c2-8debef3841b9"

# -- TransJatim --
export TTJ_UUID="112d8f49-75b5-4b55-b463-327b1d540b1b"
export RT_UUID="bcb590e8-2817-4864-8ecd-5293c73c8de0"

# 2. Substitute and output to the migration folder
envsubst < "$SCRIPT_DIR/template.sql" > "$SCRIPT_DIR/../migrations/004-seed.sql"
