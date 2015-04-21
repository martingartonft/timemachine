curl http://www.ft.com/rss/home/uk | grep ^.link.http://www.ft.com/cms/s/ | sed s,\.html\?.*,, | sed s,^.*\/,, | xargs -ixx -n1 echo \"xx\",

