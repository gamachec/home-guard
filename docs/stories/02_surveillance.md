### Story 2 : Surveillance et Terminaison de Processus

**Objectif :** Implémenter la capacité technique de scanner les processus actifs et d'en fermer certains.

* **Détails techniques :**
    * Utiliser des appels système Windows (via `golang.org/x/sys/windows`) ou une bibliothèque comme
      `shirou/gopsutil/process`.
    * Créer une fonction qui prend une liste de noms d'exécutables (ex: `["roblox.exe", "chrome.exe"]`) et vérifie s'ils
      tournent.
    * Implémenter la logique de "Kill" propre : tenter une fermeture propre d'abord, puis forcer si nécessaire.
    * **Test :** Créer un topic MQTT `cmnd/agent_pc/kill_test` qui, lorsqu'il reçoit un nom de processus, le ferme
      immédiatement.
