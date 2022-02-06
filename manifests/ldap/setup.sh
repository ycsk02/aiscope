kubectl create ns ldap
kubectl create secret generic openldap \
    --namespace ldap \
    --from-literal=adminpassword=adminpassword
kubectl create configmap ldap \
    --namespace ldap \
    --from-file=ldap/ldifs
kubectl apply --namespace ldap -f ldap/ldap.yaml

LDAP_POD=$(kubectl -n ldap get pod -l "app.kubernetes.io/name=openldap" -o jsonpath="{.items[0].metadata.name}")
kubectl -n ldap exec $LDAP_POD -- ldapadd -x -D "cn=admin,dc=aiscope,dc=io" -w adminpassword -H ldap://localhost:389 -f /ldifs/0-ous.ldif
kubectl -n ldap exec $LDAP_POD -- ldapadd -x -D "cn=admin,dc=aiscope,dc=io" -w adminpassword -H ldap://localhost:389 -f /ldifs/1-users.ldif
kubectl -n ldap exec $LDAP_POD -- ldapadd -x -D "cn=admin,dc=aiscope,dc=io" -w adminpassword -H ldap://localhost:389 -f /ldifs/2-groups.ldif
# List down the entities loaded
kubectl -n ldap exec $LDAP_POD -- \
    ldapsearch -LLL -x -H ldap://localhost:389 -D "cn=admin,dc=aiscope,dc=io" -w adminpassword -b "ou=people,dc=aiscope,dc=io" dn

kubectl -n ldap exec $LDAP_POD -- \
    ldapsearch -LLL -x -H ldap://localhost:389 -D "cn=admin,dc=aiscope,dc=io" -w adminpassword -b "ou=Users,dc=aiscope,dc=io" dn


cat << EOF > sukai.ldif
dn: uid=sukai,ou=Users,dc=aiscope,dc=io
objectClass: inetOrgPerson
objectClass: top
sn: sukai
cn: sukai
uid: sukai
mail: ycsk02@hotmail.com
userPassword: 123456
EOF
