### Story 4 : Machine à États et Logique de "Mode Limité"

**Objectif :** Gérer le cycle de vie du temps d'écran (Autorisé -> Countdown -> Bloqué).

* **Détails techniques :**
    * Définir trois états : `ACTIVE`, `WARNING`, `BLOCKED`.
    * L'agent doit écouter sur un topic `cmnd/agent_pc/mode`.
    * Si `mode == BLOCKED` : L'agent lance une goroutine qui scanne les processus interdits toutes les 2 à 5 secondes et
      les tue en boucle.
    * Si `mode == WARNING` : Déclencher la notification "5 minutes restantes".
    * L'agent doit publier son état actuel sur `stat/agent_pc/current_mode` pour que Home Assistant puisse confirmer le
      changement.
