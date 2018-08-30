## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown() {
  rio rm ${stk}
}

riorun() {
    rpt=$1
    rscl=$2
    rim=$3
    rio run -n ${stk}/${srv} -p ${rpt} --scale=${rscl} ${rim}
    rio wait ${stk}/${srv}
}

riostage() {
    sim=$1
    svs=$2
    rio stage --image=${sim} ${stk}/${srv}:${svs} 
    rio wait ${stk}/${srv}
}

rioweight() {
    wvs=$1
    wpct=$2
    rio weight ${stk}/${srv}:${wvs}=${wpct} 
}

riopromote() {
    pvs=$1
    rio promote ${stk}/${srv}:${pvs} 
}


## Validation tests ##
@test "rio run - image check" {
  tim='ibuildthecloud/demo:v1'  
  riorun '80/http' '3' "${tim}"
  got="$(rio export -o json ${stk} | jq -r '.services.'${srv}'.image')"
  expected=${tim}
  echo "Expect: ${expected}"
  echo "Got: ${got}"
  [[ "${got}" == "${expected}" ]]

}

@test "rio stage image check" {
  riorun '80/http' '3' 'ibuildthecloud/demo:v1'
  riostage 'ibuildthecloud/demo:v3' 'v3'
  got="$(rio export -o json ${stk} | jq -r '.services.'${srv}'.revisions.v3.image')"
  expected="ibuildthecloud/demo:v3"
  echo "Expect: ${expected}"
  echo "Got: ${got}"
  [[ "${got}" == "${expected}" ]]
}

@test "rio weight percentage check" {
  riorun '80/http' '3' 'ibuildthecloud/demo:v1'
  riostage 'ibuildthecloud/demo:v3' 'v3'
  rioweight 'v3' '50'
  got="$(rio export -o json ${stk} | jq -r '.services.'${srv}'.revisions.v3.weight')"
  expected="50"
  echo "Expect: ${expected}"
  echo "Got: ${got}"
  [[ "${got}" == "${expected}" ]]

}

@test "rio promote image check" {
  riorun '80/http' '3' 'ibuildthecloud/demo:v1'
  riostage 'ibuildthecloud/demo:v3' 'v3'
  rioweight 'v3' '50'
  riopromote 'v3'
  got="$(rio export -o json ${stk} | jq -r '.services.'${srv}'.image')"
  expected="ibuildthecloud/demo:v3"
  echo "Expect: ${expected}"
  echo "Got: ${got}"
  [[ "${got}" == "${expected}" ]]
  
}
