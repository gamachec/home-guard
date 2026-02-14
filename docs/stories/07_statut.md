### Story 7 : Exposer les applications lancées

**Objectif :** Exposer sur MQTT les applications actuellement lancées sur le poste (uniquement les process faisant
partie de la blacklist)

* **Détails techniques :**
    * Faire un distinct (un process peut être lancé plusieurs fois, il doit n'apparaître qu'une seule fois dans mqtt)