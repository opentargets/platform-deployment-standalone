#!/bin/bash
set -e

# check prerequisites
# an argument for the profile is provided
[ -z "$1" ] &&  echo "usage: configure <profile>" && exit 1
# the profile provided as argument exists in ./profiles
[ ! -f "./profiles/$1" ] && echo "profile $1 does not exist in ./profiles" && exit 1
# pigz must be installed
if ! command -v pigz >/dev/null 2>&1; then
  echo "pigz is not installed, please install it"
  exit 1
fi

# prepare folder and source the profile
mkdir -p deployment
cp ./profiles/"$1" ./deployment/profile
# shellcheck source=/dev/null
source ./deployment/profile


####################################################################################################
# configure opensearch
cp ./etc/opensearch.env ./deployment/opensearch.env

# set log level to DEBUG if OT_DEBUG is true
if [ "$OT_DEBUG" == "true" ]; then
  sed -i "/logger.level=\"/s/[A-Z]\+/DEBUG/" ./deployment/opensearch.env
fi


####################################################################################################
# configure api
cp ./etc/api.env ./deployment/api.env

# split OT_RELEASE and replace into env vars
IFS='.-' read -r YEAR MONTH ITERATION <<< "$OT_RELEASE"
sed -i "/META_DATA_YEAR=/s/0/$YEAR/" ./deployment/api.env
sed -i "/META_DATA_MONTH=/s/0/$MONTH/" ./deployment/api.env
sed -i "/META_DATA_ITERATION=/s/0/${ITERATION:-final}/" ./deployment/api.env

# split OT_API_TAG and replace into env vars
IFS='.-' read -r MAJOR MINOR PATCH _ <<< "$OT_API_TAG"
sed -i "/META_APIVERSION_MAJOR=/s/0/$MAJOR/" ./deployment/api.env
sed -i "/META_APIVERSION_MINOR=/s/0/$MINOR/" ./deployment/api.env
sed -i "/META_APIVERSION_PATCH=/s/0/$PATCH/" ./deployment/api.env

# disable cache if OT_DEBUG is true
if [ "$OT_DEBUG" == "true" ]; then
  sed -i "/PLATFORM_API_IGNORE_CACHE=/s/false/true/" ./deployment/api.env
fi

####################################################################################################
# configure openai
cp ./etc/openai.env ./deployment/openai.env


####################################################################################################
# configure web app
cp ./etc/webapp.env ./deployment/webapp.env

if [ "$OT_EXPOSE" == "true" ]; then
  HOST_IP=$(hostname -i | awk '{print $1}')
  sed -i "/WEBAPP_API_URL=/s/localhost/$HOST_IP/" ./deployment/webapp.env
  sed -i "/WEBAPP_OPENAI_URL=/s/localhost/$HOST_IP/" ./deployment/webapp.env
fi

####################################################################################################
touch ./deployment/.configured
echo "successfully configured $1 profile"
