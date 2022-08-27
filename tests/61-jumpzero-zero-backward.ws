jump
 
label:T	
push  1 	
output	
 	number
mark  label:S 
push  2 	 
output	
 	number
exit

mark
  label:T	
push  3 		
output	
 	number;push  0 
jump
	if zero;to label:S
push  4 	  
output	
 	number
exit
output:32
