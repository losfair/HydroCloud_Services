#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>

#define MAX_HEX_NUMBER_COUNT 8
#define ALLOWED_UID 1000

int ishexdigit(char ch) 
{
   if((ch>='0'&&ch<='9')||(ch>='a'&&ch<='f')||(ch>='A'&&ch<='F'))
      return(1);
   return(0);
}

int isIP6str(char *str)
{ 
   int hdcount=0;
   int hncount=0;
   int err=0;
   int packed=0;

   if(*str==':')
   {
      str++;    
      if(*str!=':')
         return(0);
      else
      {
         packed=1;
         hncount=1;
         str++;

         if(*str==0)
            return(1);
      }
   }

   if(ishexdigit(*str)==0)
   {
      return(0);        
   }

   hdcount=1;
   hncount=1;
   str++;

   while(err==0&&*str!=0)   
   {                      
      if(*str==':')
      {
         str++;
         if(*str==':')
         {
           if(packed==1)
              err=1;
           else
           {
              str++;

          if(ishexdigit(*str)||*str==0&&hncount<MAX_HEX_NUMBER_COUNT)
          {
             packed=1;
             hncount++;

             if(ishexdigit(*str))
             {
                if(hncount==MAX_HEX_NUMBER_COUNT)
                {
                   err=1;
                } else
                {
                   hdcount=1;
                   hncount++;
                   str++;   
                }
             }
          } else
          {
             err=1;
          }
       }
    } else
    {
           if(!ishexdigit(*str))
           {
              err=1;
           } else
           {
              if(hncount==MAX_HEX_NUMBER_COUNT)
              {
                 err=1;
              } else
              {
                  hdcount=1;
                  hncount++;
                  str++;   
              }
           }
        }
     } else
     {  
        if(ishexdigit(*str))
        {
           if(hdcount==4)
              err=1;
           else
           {
              hdcount++;          
              str++;
           }
         } else
            err=1;
     } 
   }

   if(hncount<MAX_HEX_NUMBER_COUNT&&packed==0)
      err=1;

    return(err==0);
}
int main(int argc, char *argv[]) {
	char *ip=NULL;
	char *str=NULL;
	if(argc!=2) {
		printf("Illegal arguments\n");
		return -1;
	}
	ip=argv[1];
	if(!isIP6str(ip)) {
		printf("Illegal IPv6 address\n");
		return -2;
	}
	str=malloc(64+strlen(ip));

	sprintf(str,"ip addr add \"%s\" dev lo",ip);
	
	if(getuid()!=ALLOWED_UID) {
		printf("Unauthorized UID\n");
		return -3;
	}

	setuid(0);
	execl("/sbin/ip","/sbin/ip","-6","addr","add",ip,"dev","lo",NULL);

	return 0;
}
