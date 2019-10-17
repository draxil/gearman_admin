gearman_admin
=============

Gearman admin protocol client for go / golang. Can be used to check on a gearman server status.

So far can be used to check workers (useful to check your worker processes are up) and on functions & jobs (good to catch a bottleneck or failure):

 	gma, err := gearman_admin.Connect("tcp", "10.0.50.17:4730")
  
  workers, err := gma.Workers()
  for _, worker := range(workers) {
      fmt.Println(worker.Addr, ": ", worker.ClientID)
  }
  
  func_status, err := gma.Status()
  for _, fs := range(func_status) {
      fmt.Println(fs.Name, ": ", fs.UnfinishedJobs) 
  }

[![Build Status](https://travis-ci.org/draxil/gearman_admin.png?branch=master)](https://travis-ci.org/draxil/gearman_admin)
[![GoDoc](https://godoc.org/github.com/draxil/gearman_admin?status.png)](https://godoc.org/github.com/draxil/gearman_admin)
